package config

import (
	"encoding/json"

	"github.com/dhontecillas/hfw/pkg/mailer"
)

func newMailtrapMailer(conf json.RawMessage) (mailer.Mailer, error) {
	var mtc mailer.MailtrapConfig
	err := json.Unmarshal(conf, &mtc)
	if err != nil {
		return nil, err
	}
	return mailer.NewMailtrapMailer(mtc)
}
