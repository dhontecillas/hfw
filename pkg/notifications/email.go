package notifications

import (
	"fmt"

	"github.com/dhontecillas/hfw/pkg/mailer"
)

// EmailCarrier contains the data to send emails
type EmailCarrier struct {
	Mailer mailer.Mailer
}

// NewEmailCarrier creates a new email carrier using a given Mailer
func NewEmailCarrier(mailSender mailer.Mailer) *EmailCarrier {
	return &EmailCarrier{
		Mailer: mailSender,
	}
}

// Send sends the email
func (c *EmailCarrier) Send(content *ContentSet, data map[string]interface{}) error {
	toAddress, ok := data["to_address"].(string)
	if !ok {
		return fmt.Errorf("missing 'to_address' field")
	}
	subject, ok := content.Texts["subject"]
	if !ok {
		return fmt.Errorf("missing 'subject' content")
	}
	textBody, ok := content.Texts["content"]
	if !ok {
		return fmt.Errorf("missing 'text body' content")
	}
	htmlBody, ok := content.HTMLs["content"]
	if !ok {
		return fmt.Errorf("missing 'html body' content")
	}

	toName, ok := data["to_name"].(string)
	if !ok {
		toName = toAddress
	}

	fromAddress, fromName := c.Mailer.Sender()
	if name, ok := data["from_name"].(string); ok {
		fromName = name
	}
	if address, ok := data["from_address"].(string); ok {
		fromAddress = address
	}

	emailMessage := mailer.Email{
		To: mailer.User{
			Name:    toName,
			Address: toAddress,
		},
		From: mailer.User{
			Name:    fromName,
			Address: fromAddress,
		},
		Subject: subject,
		HTML:    htmlBody,
		Text:    textBody,
	}

	return c.Mailer.Send(emailMessage)
}
