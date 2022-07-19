package config

import (
	"github.com/spf13/viper"

	"github.com/dhontecillas/hfw/pkg/mailer"
)

const (
	confKeyMailtrapUser     string = "mailtrap.user"
	confKeyMailtrapPassword string = "mailtrap.password"
)

func newMailtrapMailer(confPrefix string) (mailer.Mailer, error) {
	return mailer.NewMailtrapMailer(mailer.MailtrapConfig{
		User:     viper.GetString(confPrefix + confKeyMailtrapUser),
		Password: viper.GetString(confPrefix + confKeyMailtrapPassword),
	})
}
