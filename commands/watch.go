package commands

import (
	"context"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/fsnotify/fsnotify"
	"github.com/maliceio/malice/config"
)

func cmdWatch(folderName string, logs bool) error {

	log.WithFields(log.Fields{
		"env": config.Conf.Environment.Run,
	}).Info("Malice watching folder: ", folderName)

	info, err := os.Stat(folderName)

	// Check that folder exists
	if os.IsNotExist(err) {
		log.Error("error: folder does not exist.")
		return nil
	}
	// Check that path is a folder and not a file
	if info.IsDir() {
		if err := NewWatcher(folderName); err != nil {
			return err
		}
	} else {
		log.Error("error: path is not a folder")
	}

	return nil
}

// NewWatcher creates a new watcher for the user supplied folder with proper graceful shutdown
func NewWatcher(folder string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	defer close(done)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("watcher goroutine panic: %v", r)
				done <- nil
			}
		}()

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.WithField("event", event).Debug("file system event")
				if event.Op&fsnotify.Create == fsnotify.Create {
					log.WithField("file", event.Name).Debug("file created, scanning")
					// Scan new sample in watch folder
					if err := cmdScan(event.Name, false); err != nil {
						log.WithError(err).Error("scan failed")
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.WithError(err).Error("watcher error")
				done <- err
				return
			case <-ctx.Done():
				log.Debug("watcher shutting down")
				done <- nil
				return
			}
		}
	}()

	if err = watcher.Add(folder); err != nil {
		return err
	}

	// Block until error or context cancellation
	return <-done
}
