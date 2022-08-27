package mailer

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/mailgun/mailgun-go/v4"
)

// MailgunConfig has the parameters to use the Mailgun service
type MailgunConfig struct {
	Domain      string
	Key         string
	UseEUServer bool
}

// MailgunMailer implements the Mailer interface to send emails using Mailgun service
type MailgunMailer struct {
	client *mailgun.MailgunImpl
}

// NewMailgunMailer creates a new Mailer to send emails through Mailgun
func NewMailgunMailer(conf MailgunConfig) (*MailgunMailer, error) {
	if len(conf.Key) == 0 {
		return nil, fmt.Errorf("missing Mailgun Key")
	}
	if len(conf.Domain) == 0 {
		return nil, fmt.Errorf("missing Mailgun Domain")
	}
	client := mailgun.NewMailgun(conf.Domain, conf.Key)
	if conf.UseEUServer {
		client.SetAPIBase(mailgun.APIBaseEU)
	}
	return &MailgunMailer{
		client: client,
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
