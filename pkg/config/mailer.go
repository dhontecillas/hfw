package config

import (
	"fmt"

	"github.com/spf13/viper"

	"github.com/dhontecillas/hfw/pkg/mailer"
	"github.com/dhontecillas/hfw/pkg/obs"
)

const (
	sendgridMailer  string = "sendgrid"
	mailtrapMailer  string = "mailtrap"
	roundCubeMailer string = "roundcube"
	consoleMailer   string = "console"
	nopMailer       string = "nop"

	confKeyMailerPreferred string = "mailer.preferred"
	confKeyMailerLogs      string = "mailer.logs"
)

// MailerConfig contains the selected mailer configuration
type MailerConfig struct {
	Name          string
	LogSentEmails bool
	ConfPrefix    string
}

func (m *MailerConfig) String() string {
	return fmt.Sprintf("Mailer %s (log enabled: %t)", m.Name, m.LogSentEmails)
}

// ReadMailerConfig returns the name of a mailer and a boolean to select
// if logging should be enabled
func ReadMailerConfig(ins *obs.Insighter, confPrefix string) (*MailerConfig, error) {
	// Allow to override mailer selection with preferredmailer and
	// preferredmailerlogs
	confKey := confPrefix + confKeyMailerPreferred
	selectedMailer := viper.GetString(confKey)
	if len(selectedMailer) == 0 {
		return nil, fmt.Errorf("cannot read preferred mailer from %s",
			confKey)
	}

	allowedValues := map[string]bool{
		sendgridMailer:  true,
		mailtrapMailer:  true,
		roundCubeMailer: true,
		consoleMailer:   true,
		nopMailer:       true,
	}
	if _, ok := allowedValues[selectedMailer]; !ok {
		msg := fmt.Sprintf("cannot find mailer: %s", selectedMailer)
		ins.L.Panic(msg)
		panic(msg)
	}

	return &MailerConfig{
		Name:          selectedMailer,
		LogSentEmails: viper.GetBool(confPrefix + confKeyMailerLogs),
		ConfPrefix:    confPrefix,
	}, nil
}

// CreateMailer creates a mailer from a provided MailerConfig
func CreateMailer(ins *obs.Insighter, mailerConf *MailerConfig) (mailer.Mailer, error) {
	if mailerConf == nil {
		err := fmt.Errorf("no mailerConf provided")
		ins.L.Err(err, "cannot create mailer")
		return nil, err
	}
	// check configuration and use approppriate mailer
	var m mailer.Mailer
	var err error

	infMsg := fmt.Sprintf("Creating mailer: %s\n", mailerConf.String())
	ins.L.Info(infMsg)
	switch mailerConf.Name {
	case consoleMailer:
		m = mailer.NewConsoleMailer()
	case mailtrapMailer:
		m, err = newMailtrapMailer(mailerConf.ConfPrefix)
	case roundCubeMailer:
		m = mailer.NewRoundcubeMailer()
	case nopMailer:
		m = mailer.NewNopMailer()
	default:
		m, err = newSendgridMailer(mailerConf.ConfPrefix)
	}

	if err != nil {
		ins.L.Err(err, "cannot create mailer")
		return nil, err
	}
	if mailerConf.LogSentEmails {
		m = mailer.NewLoggerMailer(m, ins)
	}
	return m, nil
}
