package mailer

import (
	"fmt"
	"net/smtp"
)

// MailtrapConfig has the configurations to use the Mailtrap service
type MailtrapConfig struct {
	Server   string
	Port     int
	User     string
	Password string
}

type mailtrapMailer struct {
	Server   string
	Port     int
	User     string
	Password string
}

// NewMailtrapMailer instantiates a new Mailtrap mailer.
func NewMailtrapMailer(conf MailtrapConfig) (Mailer, error) {
	if len(conf.User) == 0 {
		return nil, fmt.Errorf("missing Mailtrap user")
	}
	if len(conf.Password) == 0 {
		return nil, fmt.Errorf("missing Mailtrap password")
	}
	// for server and port, we can fall back to defaults:
	if len(conf.Server) == 0 {
		conf.Server = "smtp.mailtrap.io"
	}
	if conf.Port == 0 {
		conf.Port = 587
	}
	return &mailtrapMailer{
		Server:   conf.Server,
		Port:     conf.Port,
		User:     conf.User,
		Password: conf.Password,
	}, nil
}

// Sends an email to Mailtrap service
func (m *mailtrapMailer) Send(e Email) error {
	smtpServer := fmt.Sprintf("%s:%d", m.Server, m.Port)
	auth := smtp.PlainAuth("", m.User, m.Password, m.Server)
	msg := ComposeSMTPMsg(e)
	err := smtp.SendMail(smtpServer, auth, e.From.Address, []string{e.To.Address}, []byte(msg))
	return err
}

// Sender returns the default sender address and name
func (m *mailtrapMailer) Sender() (string, string) {
	return "noreply@example.com", "No Reply"
}
