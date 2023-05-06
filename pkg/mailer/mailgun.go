package mailer

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/mailgun/mailgun-go/v4"
)

// MailgunMailer implements the Mailer interface to send emails using Mailgun service
type MailgunMailer struct {
	client *mailgun.MailgunImpl

	senderEmail string
	senderName  string
}

// NewMailgunMailer creates a new Mailer to send emails through Mailgun
func NewMailgunMailer(domain string, key string, senderEmail string,
	senderName string, useEUServer bool) (*MailgunMailer, error) {

	if len(domain) == 0 {
		return nil, fmt.Errorf("missing Mailgun Domain")
	}
	if len(key) == 0 {
		return nil, fmt.Errorf("missing Mailgun Key")
	}
	if len(senderEmail) == 0 {
		return nil, fmt.Errorf("missing Mailgun Sender Email Address")
	}
	if len(senderName) == 0 {
		senderName = senderEmail
	}

	client := mailgun.NewMailgun(domain, key)
	if useEUServer {
		client.SetAPIBase(mailgun.APIBaseEU)
	}
	return &MailgunMailer{
		client:      client,
		senderEmail: senderEmail,
		senderName:  senderName,
	}, nil
}

// Send sends an email through Mailgun
func (m *MailgunMailer) Send(e Email) error {
	message := m.client.NewMessage(e.From.Address, e.Subject, e.Text, e.To.Address)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Send the message with a 10 second timeout
	// we ignore the "msg" and "id" for the email
	_, _, err := m.client.Send(ctx, message)

	if err != nil {
		return errors.Wrap(err, "error sending email")
	}
	return nil
}

// Sender returns the default sender address and name
func (m *MailgunMailer) Sender() (string, string) {
	return m.senderEmail, m.senderName
}
