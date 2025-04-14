package config

import (
	"encoding/json"
	"fmt"

	"github.com/dhontecillas/hfw/pkg/mailer"
	"github.com/dhontecillas/hfw/pkg/obs"
)

const (
	sendgridMailer  string = "sendgrid"
	mailtrapMailer  string = "mailtrap"
	roundCubeMailer string = "roundcube"
	consoleMailer   string = "console"
	nopMailer       string = "nop"

	confKeyMailer            string = "mailer"
	confKeyMailerPreferred   string = "mailer.preferred"
	confKeyMailerLogs        string = "mailer.logs"
	confKeyMailerFromAddress string = "mailer.from.address"
	confKeyMailerFromName    string = "mailer.from.name"
)

type MailerAddressConfig struct {
	Address string `json:"from_address"`
	Name    string `json:"from_name"`
}

// MailerConfig contains the selected mailer configuration
type MailerConfig struct {
	Name          string              `json:"preferred"`
	LogSentEmails bool                `json:"logs"`
	From          MailerAddressConfig `json:"from"`
	Config        json.RawMessage     `json:"config"`
}

func (m *MailerConfig) String() string {
	return fmt.Sprintf("Mailer %s (log enabled: %t) %s <%s>",
		m.Name, m.LogSentEmails, m.From.Name, m.From.Address)
}

// ReadMailerConfig returns the name of a mailer and a boolean to select
// if logging should be enabled
func ReadMailerConfig(ins *obs.Insighter, conf ConfLoader) (*MailerConfig, error) {
	conf, err := conf.Section([]string{confKeyMailer})
	if err != nil {
		// TODO: improve error
		return nil, err
	}
	var mailerConf MailerConfig
	err = conf.Parse(&mailerConf)
	if err != nil {
		return nil, err
	}

	// Allow to override mailer selection with preferredmailer and
	// preferredmailerlogs
	if mailerConf.Name == "" {
		return nil, fmt.Errorf("empty preferred mailer")
	}

	allowedValues := map[string]bool{
		sendgridMailer:  true,
		mailtrapMailer:  true,
		roundCubeMailer: true,
		consoleMailer:   true,
		nopMailer:       true,
	}
	if _, ok := allowedValues[mailerConf.Name]; !ok {
		err := fmt.Errorf("cannot find mailer: %s", mailerConf.Name)
		return nil, err
	}
	if mailerConf.From.Address == "" {
		// TODO: check with a regex that is a valid email address
		err := fmt.Errorf("empty from address")
		return nil, err
	}
	if mailerConf.From.Name == "" {
		mailerConf.From.Name = mailerConf.From.Address
	}
	return &mailerConf, nil
}

// CreateMailer creates a mailer from a provided MailerConfig
func CreateMailer(ins *obs.Insighter, mailerConf *MailerConfig) (mailer.Mailer, error) {
	if mailerConf == nil {
		err := fmt.Errorf("no mailerConf provided")
		ins.L.Err(err, "cannot create mailer", map[string]interface{}{
			"conf": mailerConf.String(),
		})
		return nil, err
	}
	// check configuration and use approppriate mailer
	var m mailer.Mailer
	var err error

	ins.L.Info("Creating mailer: %s\n", map[string]interface{}{
		"conf": mailerConf.String(),
	})
	switch mailerConf.Name {
	case consoleMailer:
		m = mailer.NewConsoleMailer()
	case mailtrapMailer:
		m, err = newMailtrapMailer(mailerConf.Config)
	case roundCubeMailer:
		m = mailer.NewRoundcubeMailer()
	case nopMailer:
		m = mailer.NewNopMailer()
	default:
		// TODO: sendgrid MUST repeat the from address inside the config
		m, err = newSendgridMailer(ins, mailerConf.Config)
	}

	if err != nil {
		ins.L.Err(err, "cannot create mailer", nil)
		return nil, err
	}
	if mailerConf.LogSentEmails {
		m = mailer.NewLoggerMailer(m, ins)
	}
	return m, nil
}
