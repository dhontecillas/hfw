package mailer

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// SendGridConfig has the parameters to use the SendGrid service
type SendGridConfig struct {
	Key string
}

// SendGridMailer implements the Mailer interface to send emails using SendGrid service
type SendGridMailer struct {
	client *sendgrid.Client
}

// NewSendGridMailer creates a new Mailer to send emails through SendGrid
func NewSendGridMailer(conf SendGridConfig) (*SendGridMailer, error) {
	if len(conf.Key) == 0 {
		return nil, fmt.Errorf("missing SendGrid key")
	}
	return &SendGridMailer{
		client: sendgrid.NewSendClient(conf.Key),
	}, nil
}

// Send sends an email through SendGrid
func (m *SendGridMailer) Send(e Email) error {
	resp, err := m.client.Send(
		mail.NewSingleEmail(
			mail.NewEmail(e.From.Name, e.From.Address),
			e.Subject,
			mail.NewEmail(e.To.Name, e.To.Address),
			e.Text,
			e.HTML))
	if err != nil {
		return errors.Wrap(err, "error sending email")
	}
	if resp.StatusCode >= 400 {
		return errors.New(fmt.Sprintf("error sending email: %s", resp.Body))
	}
	return nil
}
