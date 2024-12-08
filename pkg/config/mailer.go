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

	confKeyMailerPreferred   string = "mailer.preferred"
	confKeyMailerLogs        string = "mailer.logs"
	confKeyMailerFromAddress string = "mailer.from.address"
	confKeyMailerFromName    string = "mailer.from.name"
)

// MailerConfig contains the selected mailer configuration
type MailerConfig struct {
	Name          string `json:"name"`
	LogSentEmails bool   `json:"log_sent_mails"`
	ConfPrefix    string `json:"conf_prefix"`
	FromAddress   string `json:"from_address"`
	FromName      string `json:"from_name"`
}

func (m *MailerConfig) String() string {
	return fmt.Sprintf("Mailer %s (log enabled: %t) %s <%s>",
		m.Name, m.LogSentEmails, m.FromName, m.FromAddress)
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
		ins.L.Panic("cannot find mailer", map[string]interface{}{
			"mailer": selectedMailer,
		})
		panic(msg)
	}

	confKey = confPrefix + confKeyMailerFromAddress
	if !viper.IsSet(confKey) {
		msg := fmt.Sprintf("cannot read mailer sender address: %s", confKey)
		ins.L.Panic(msg, nil)
		panic(msg)
	}
	fromAddress := viper.GetString(confKey)
	fromName := viper.GetString(confPrefix + confKeyMailerFromName)
	if fromName == "" {
		fromName = fromAddress
	}

	return &MailerConfig{
		Name:          selectedMailer,
		LogSentEmails: viper.GetBool(confPrefix + confKeyMailerLogs),
		ConfPrefix:    confPrefix,
		FromAddress:   fromAddress,
		FromName:      fromName,
	}, nil
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
		m, err = newMailtrapMailer(mailerConf.ConfPrefix)
	case roundCubeMailer:
		m = mailer.NewRoundcubeMailer()
	case nopMailer:
		m = mailer.NewNopMailer()
	default:
		m, err = newSendgridMailer(ins, mailerConf.ConfPrefix,
			mailerConf.FromAddress, mailerConf.FromName)
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
