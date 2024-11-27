package config

import (
	"fmt"

	"github.com/dhontecillas/hfw/pkg/mailer"
	"github.com/dhontecillas/hfw/pkg/obs"

	"github.com/spf13/viper"
)

const (
	confKeySendgridKey         string = "sendgrid.key"
	confKeySendgridSenderEmail string = "sendgrid.senderemail"
	confKeySendgridSenderName  string = "sendgrid.sendername"
)

// SendGridConfig has the parameters to use the SendGrid service
type SendGridConfig struct {
	Key string
}

// configSendGrid fills a SendGridConfig struct from viper parameters
func configSendGrid(confPrefix string) (SendGridConfig, error) {
	if !viper.IsSet(confPrefix + confKeySendgridKey) {
		return SendGridConfig{}, fmt.Errorf("missing sendgrid Key: %s",
			confPrefix+confKeySendgridKey)
	}
	key := viper.GetString(confPrefix + confKeySendgridKey)

	return SendGridConfig{
		Key: key,
	}, nil
}

func newSendgridMailer(ins *obs.Insighter, confPrefix string,
	from string, name string) (mailer.Mailer, error) {
	conf, err := configSendGrid(confPrefix)
	if err != nil {
		return nil, err
	}
	ins.L.Info(fmt.Sprintf("new sendgrid config %#v", conf), nil)
	m, err := mailer.NewSendGridMailer(conf.Key, from, name)
	ins.L.Info(fmt.Sprintf("created mailer: %#v", m), nil)
	return m, err
}
