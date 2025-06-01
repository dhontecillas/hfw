package config

import (
	"encoding/json"

	"github.com/dhontecillas/hfw/pkg/mailer"
	"github.com/dhontecillas/hfw/pkg/obs"
)

// SendGridConfig has the parameters to use the SendGrid service
type SendGridConfig struct {
	Key         string `json:"key"`
	SenderEmail string `json:"senderemail"`
	SenderName  string `json:"sendername"`
}

func configSendGrid(conf json.RawMessage) (SendGridConfig, error) {
	var sgConf SendGridConfig
	err := json.Unmarshal(conf, &sgConf)
	if err != nil {
		return sgConf, err
	}
	return sgConf, nil
}

func newSendgridMailer(ins *obs.Insighter, conf json.RawMessage) (mailer.Mailer, error) {
	sgConf, err := configSendGrid(conf)
	if err != nil {
		ins.L.Err(err, "cannot read sendgrid config", nil)
	}
	m, err := mailer.NewSendGridMailer(sgConf.Key, sgConf.SenderEmail, sgConf.SenderName)
	if err != nil {
		ins.L.Err(err, "cannot create sendgrid mailer", nil)
	}
	// ins.L.Info(fmt.Sprintf("created mailer: %#v", m), nil)
	return m, err
}
