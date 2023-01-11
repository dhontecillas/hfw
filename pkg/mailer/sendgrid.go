package mailer

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// SendGridConfig has the parameters to use the SendGrid service
type SendGridConfig struct {
	Key         string
	FromAddress string
	FromName    string
}

// SendGridMailer implements the Mailer interface to send emails using SendGrid service
type SendGridMailer struct {
	client      *sendgrid.Client
	FromAddress string
	FromName    string
}

// NewSendGridMailer creates a new Mailer to send emails through SendGrid
func NewSendGridMailer(conf SendGridConfig) (*SendGridMailer, error) {
	if len(conf.Key) == 0 {
		return nil, fmt.Errorf("missing SendGrid key")
	}
	if len(conf.FromAddress) == 0 {
		return nil, fmt.Errorf("missing from address")
	}
	name := conf.FromName
	if len(name) == 0 {
		name = conf.FromAddress
	}
	return &SendGridMailer{
		client:      sendgrid.NewSendClient(conf.Key),
		FromAddress: conf.FromAddress,
		FromName:    name,
	}, nil
}

// Send sends an email through SendGrid
func (m *SendGridMailer) Send(e Email) error {
	resp, err := m.client.Send(
		mail.NewSingleEmail(
			mail.NewEmail(m.FromName, m.FromAddress),
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

func (m *SendGridMailer) Sender() (string, string) {
	return m.FromAddress, m.FromName
}
