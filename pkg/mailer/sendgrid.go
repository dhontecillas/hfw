package mailer

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// SendGridMailer implements the Mailer interface to send emails using SendGrid service
type SendGridMailer struct {
	client *sendgrid.Client

	fromAddress string
	fromName    string
}

// NewSendGridMailer creates a new Mailer to send emails through SendGrid
func NewSendGridMailer(key, fromAddress, fromName string) (*SendGridMailer, error) {
	if len(key) == 0 {
		return nil, fmt.Errorf("missing SendGrid key")
	}
	if len(fromAddress) == 0 {
		return nil, fmt.Errorf("missing from address")
	}
	if len(fromName) == 0 {
		fromName = fromAddress
	}
	return &SendGridMailer{
		client:      sendgrid.NewSendClient(key),
		fromAddress: fromAddress,
		fromName:    fromName,
	}, nil
}

// Send sends an email through SendGrid
func (m *SendGridMailer) Send(e Email) error {
	// TODO: we can check that e.From user matches the
	// configured sendgrid from sender, and emit a warning
	// on mismatch
	resp, err := m.client.Send(
		mail.NewSingleEmail(
			mail.NewEmail(m.fromName, m.fromAddress),
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
	return m.fromAddress, m.fromName
}
