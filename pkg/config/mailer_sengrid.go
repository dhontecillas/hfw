package config

import (
	"fmt"

	"github.com/dhontecillas/hfw/pkg/mailer"
	"github.com/dhontecillas/hfw/pkg/obs"

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

func newSendgridMailer(ins *obs.Insighter, confPrefix string,
	from string, name string) (mailer.Mailer, error) {

	conf, err := configSendGrid(confPrefix)
	if err != nil {
		return nil, err
	}
	conf.FromAddress = from
	conf.FromName = name
	ins.L.Info(fmt.Sprintf("new sendgrid config %#v", conf))
	m, err := mailer.NewSendGridMailer(conf)
	ins.L.Info(fmt.Sprintf("created mailer: %#v", m))
	return m, err
}
