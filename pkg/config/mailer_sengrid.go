package config

import (
	"fmt"

	"github.com/dhontecillas/hfw/pkg/mailer"
	"github.com/spf13/viper"
)

const confKeySendgridKey string = "sendgrid.key"

// configSendGrid fills a SendGridConfig struct from viper parameters
func configSendGrid(confPrefix string) (mailer.SendGridConfig, error) {
	if !viper.IsSet(confPrefix + confKeySendgridKey) {
		return mailer.SendGridConfig{}, fmt.Errorf("missing sendgrid Key: %s",
			confPrefix+confKeySendgridKey)
	}
	val := viper.GetString(confPrefix + confKeySendgridKey)
	return mailer.SendGridConfig{
		Key: val,
	}, nil
}

func newSendgridMailer(confPrefix string) (mailer.Mailer, error) {
	conf, err := configSendGrid(confPrefix)
	if err != nil {
		return nil, err
	}
	return mailer.NewSendGridMailer(conf)
}
