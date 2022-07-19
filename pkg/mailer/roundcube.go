package mailer

import (
	"fmt"
	"net/smtp"
)

// RoundcubeMailer is an SMTP server for local dev-env environments.
type RoundcubeMailer struct {
	Server string
	Port   int
}

// NewRoundcubeMailer instantiates a new Roundcube mailer.
func NewRoundcubeMailer() *RoundcubeMailer {
	return &RoundcubeMailer{
		Server: "mail",
		Port:   1025,
	}
}

// Send sends an email to Roundcube service.
func (m *RoundcubeMailer) Send(e Email) error {
	smtpServer := fmt.Sprintf("%s:%d", m.Server, m.Port)
	msg := ComposeSMTPMsg(e)
	err := smtp.SendMail(smtpServer, nil, e.From.Address, []string{e.To.Address}, []byte(msg))
	return err
}
