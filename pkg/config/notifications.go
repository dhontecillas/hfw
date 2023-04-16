package config

import (
	"fmt"

	"github.com/dhontecillas/hfw/pkg/mailer"
	"github.com/dhontecillas/hfw/pkg/notifications"
	"github.com/dhontecillas/hfw/pkg/obs"
	"github.com/spf13/viper"
)

const (
	confKeyNotificationsTemplatesDir string = "notifications.templates.dir"
)

// NotificationsConfig contains the configuration for sending
// notifications.
type NotificationsConfig struct {
	NotificationsTemplatesDir string
}

// ReadNotificationsConfig creates a NotificationsConfig instance from
// the existing environment or config file values.
func ReadNotificationsConfig(ins *obs.Insighter, confPrefix string) (*NotificationsConfig, error) {
	notificationsTemplatesDir := viper.GetString(
		confPrefix + confKeyNotificationsTemplatesDir)
	if len(notificationsTemplatesDir) == 0 {
		return nil, fmt.Errorf("cannot read preferred mailer")
	}

	return &NotificationsConfig{
		NotificationsTemplatesDir: notificationsTemplatesDir,
	}, nil
}

// CreateNotificationsComposer create a notifications.Notifier from a provided
// configuration.
func CreateNotificationsComposer(ins *obs.Insighter,
	notificationsConf *NotificationsConfig,
	mailer mailer.Mailer) (notifications.Composer, error) {

	return notifications.NewFileSystemComposer(notificationsConf.NotificationsTemplatesDir), nil
}
