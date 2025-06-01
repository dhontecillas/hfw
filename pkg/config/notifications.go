package config

import (
	"fmt"

	"github.com/dhontecillas/hfw/pkg/mailer"
	"github.com/dhontecillas/hfw/pkg/notifications"
	"github.com/dhontecillas/hfw/pkg/obs"
)

const (
	confKeyNotificationsTemplatesDir string = "notifications.templates.dir"
)

// NotificationsConfig contains the configuration for sending
// notifications.
type NotificationsConfig struct {
	NotificationsTemplatesDir string `json:"dir"`
}

// ReadNotificationsConfig creates a NotificationsConfig instance from
// the existing environment or config file values.
func ReadNotificationsConfig(ins *obs.Insighter, cldr ConfLoader) (*NotificationsConfig, error) {
	cldr, err := cldr.Section([]string{"notifications", "templates"})
	if err != nil {
		return nil, err
	}
	var notCfg NotificationsConfig
	err = cldr.Parse(&notCfg)
	if err != nil {
		return nil, err
	}

	if notCfg.NotificationsTemplatesDir == "" {
		return nil, fmt.Errorf("cannot read notifications templates dir")
	}

	return &notCfg, nil
}

// CreateNotificationsComposer create a notifications.Notifier from a provided
// configuration.
func CreateNotificationsComposer(ins *obs.Insighter,
	notificationsConf *NotificationsConfig,
	mailer mailer.Mailer) (notifications.Composer, error) {

	return notifications.NewFileSystemComposer(notificationsConf.NotificationsTemplatesDir), nil
}
