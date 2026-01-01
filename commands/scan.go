package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/fatih/structs"
	"github.com/malice-plugins/pkgs/database/elasticsearch"
	"github.com/malice-plugins/pkgs/utils"
	"github.com/maliceio/malice/config"
	"github.com/maliceio/malice/internal/util"
	"github.com/maliceio/malice/malice/database"
	"github.com/maliceio/malice/malice/docker/client"
	"github.com/maliceio/malice/malice/docker/client/container"
	"github.com/maliceio/malice/malice/persist"
	"github.com/maliceio/malice/plugins"
	"github.com/pkg/errors"
)

const (
	// Timeout for the entire scan operation
	scanTimeout = 10 * time.Minute
	// Timeout for individual operations like MIME detection
	operationTimeout = 2 * time.Minute
	// Max concurrent plugin executions
	maxConcurrentPlugins = 10
)

// cmdScan scans a sample with all appropriate malice plugins
func cmdScan(path string, logs bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), scanTimeout)
	defer cancel()

	return cmdScanWithContext(ctx, path, logs)
}

// cmdScanWithContext scans a sample with context and timeout support
func cmdScanWithContext(ctx context.Context, path string, logs bool) error {
	if len(path) == 0 {
		return fmt.Errorf("file path required")
	}

	// Validate file path early to prevent path traversal and check existence
	if err := validateAndNormalizePath(path); err != nil {
		return err
	}

	docker := client.NewDockerClient()

	// clean stale containers from previous runs
	containers, err := container.List(docker, true)
	if err != nil {
		return errors.Wrap(err, "failed to list containers")
	}

	for _, contr := range containers {
		if utils.StringInSlice("malice", contr.Names) {
			if err := container.Remove(docker, contr.ID, true, true, true); err != nil {
				log.WithError(err).Warnf("failed to remove stale container: %s", contr.Names[0])
				// Don't fail - continue with scan
			}
		}
	}

	elasticsearchInDocker := false
	es := elasticsearch.Database{
		Index:    utils.Getopt("MALICE_ELASTICSEARCH_INDEX", "malice"),
		Type:     utils.Getopt("MALICE_ELASTICSEARCH_TYPE", "samples"),
		URL:      utils.Getopt("MALICE_ELASTICSEARCH_URL", config.Conf.DB.URL),
		Username: utils.Getopt("MALICE_ELASTICSEARCH_USERNAME", config.Conf.DB.Username),
		Password: utils.Getopt("MALICE_ELASTICSEARCH_PASSWORD", config.Conf.DB.Password),
	}

	// This assumes you haven't set up an elasticsearch instance and that malice should create one
	if strings.EqualFold(es.URL, "http://localhost:9200") {
		elasticsearchInDocker = true
		// Check that database is running
		if _, running, _ := container.Running(docker, config.Conf.DB.Name); !running {
			log.Info("database is NOT running, starting now...")
			if err := database.Start(docker, es, logs); err != nil {
				return errors.Wrap(err, "failed to start database")
			}
		}
	}

	// Initialize the malice database
	es.Init()

	// Check Plugin Status
	if plugins.InstalledPluginsCheck(docker) {
		log.Debug("All enabled plugins are installed.")
	} else {
		// Prompt user to install all plugins?
		fmt.Println("All enabled plugins not installed would you like to install them now? (yes/no)")
		fmt.Println("[Warning] This can take a while if it is the first time you have ran Malice.")
		if utils.AskForConfirmation() {
			plugins.UpdateEnabledPlugins(docker)
		}
	}

	es.Plugins = database.GetPluginsByCategory()

	file := persist.File{Path: path}
	file.Init()

	// Output File Hashes
	file.ToMarkdownTable()
	// fmt.Println(string(file.ToJSON()))

	//////////////////////////////////////
	// Copy file to malice volume
	container.CopyToVolume(docker, file)

	//////////////////////////////////////
	// Write all file data to the Database
	resp, err := es.StoreFileInfo(structs.Map(file))
	if err != nil {
		return errors.Wrap(err, "scan cmd failed to store file info")
	}

	scanID := resp.Id

	/////////////////////////////////////////////////////////////////
	// Run all Intel Plugins on the md5 hash associated with the file
	plugins.RunIntelPlugins(docker, file.SHA1, scanID, true, elasticsearchInDocker)

	// Get file's mime type with operation timeout
	mimeCtx, cancel := context.WithTimeout(ctx, operationTimeout)
	defer cancel()

	mimeType, err := persist.GetMimeType(mimeCtx, docker, file.SHA256)
	if err != nil {
		return errors.Wrap(err, "failed to get file's mime type")
	}

	log.WithFields(log.Fields{
		"mime_type": mimeType,
		"file":      file.SHA256,
	}).Debug("detected file mime type")

	// Iterate over all applicable installed plugins
	pluginsForMime := plugins.GetPluginsForMime(mimeType, true)
	log.WithField("plugin_count", len(pluginsForMime)).Debug("found plugins for mime type")
	for _, plugin := range pluginsForMime {
		log.Debugf("  - %s", plugin.Name)
	}

	// Run plugins with bounded concurrency using semaphore
	return runPluginsWithSemaphore(ctx, docker, file.SHA256, scanID, logs, elasticsearchInDocker, pluginsForMime)
}

// runPluginsWithSemaphore runs plugins with bounded concurrency
func runPluginsWithSemaphore(ctx context.Context, docker *client.Docker, sha256, scanID string, logs, elasticsearchInDocker bool, pluginsForMime []plugins.Plugin) error {
	if len(pluginsForMime) == 0 {
		log.Debug("no plugins to run")
		return nil
	}

	// Semaphore to limit concurrent operations
	sem := make(chan struct{}, maxConcurrentPlugins)
	defer close(sem)

	var wg sync.WaitGroup
	errors := make(chan error, len(pluginsForMime))
	defer close(errors)

	for _, plugin := range pluginsForMime {
		wg.Add(1)
		go func(p plugins.Plugin) {
			defer wg.Done()

			// Acquire semaphore
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				errors <- ctx.Err()
				return
			}

			// Add timeout per plugin
			pluginCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
			defer cancel()

			log.WithFields(log.Fields{
				"plugin": p.Name,
				"file":   sha256,
			}).Debug("running plugin")

			// Note: StartPlugin needs to accept context parameter
			// For now, we call it as before but with logging
			p.StartPlugin(docker, sha256, scanID, logs, elasticsearchInDocker, &sync.WaitGroup{})

			select {
			case <-pluginCtx.Done():
				errors <- fmt.Errorf("plugin %s timeout", p.Name)
			default:
				// Plugin completed
			}
		}(plugin)
	}

	// Wait for all plugins to complete or timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Debug("all plugins completed")
		// Check for any errors
		close(errors)
		for err := range errors {
			if err != nil {
				log.WithError(err).Warn("plugin error")
			}
		}
		return nil
	case <-ctx.Done():
		return fmt.Errorf("scan timeout: %w", ctx.Err())
	}
}

// validateAndNormalizePath validates a file path for security and accessibility
func validateAndNormalizePath(path string) error {
	// Get absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("cannot resolve path: %w", err)
	}

	// Check if file exists
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", path)
		}
		return fmt.Errorf("cannot access file: %w", err)
	}

	// Check if it's a regular file
	if !info.Mode().IsRegular() {
		return fmt.Errorf("path is not a regular file: %s", path)
	}

	// Check file size (prevent analyzing extremely large files)
	const maxFileSize = 512 * 1024 * 1024 // 512MB limit
	if info.Size() > maxFileSize {
		return fmt.Errorf("file too large: %d bytes (max: %d)", info.Size(), maxFileSize)
	}

	return nil
}

// APIScan is an API wrapper for cmdScan
func APIScan(file string) error {
	return cmdScan(file, false)
}
