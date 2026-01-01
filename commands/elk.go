package commands

import (
	log "github.com/Sirupsen/logrus"
	"github.com/malice-plugins/pkgs/database/elasticsearch"
	"github.com/maliceio/malice/config"
	"github.com/maliceio/malice/malice/database"
	"github.com/maliceio/malice/malice/docker/client"
	"github.com/maliceio/malice/malice/docker/client/container"
	"github.com/maliceio/malice/malice/ui"
	"github.com/maliceio/malice/malice/errors"
)

func cmdELK(logs bool) error {

	docker := client.NewDockerClient()

	if _, running, _ := container.Running(docker, config.Conf.DB.Name); !running {
		err := database.Start(docker, elasticsearch.Database{URL: config.Conf.DB.URL}, logs)
		if err != nil {
			return errors.NewScanError("elk-startup", "db_start_failed", "failed to start database", err)
		}
	} else {
		log.Warnf("container %s is already running", config.Conf.DB.Name)
	}

	if _, running, _ := container.Running(docker, config.Conf.UI.Name); !running {
		_, err := ui.Start(docker, logs)
		if err != nil {
			return errors.NewScanError("elk-startup", "ui_start_failed", "failed to start UI", err)
		}
	} else {
		log.Warnf("container %s is already running", config.Conf.UI.Name)
	}

	return nil
}
