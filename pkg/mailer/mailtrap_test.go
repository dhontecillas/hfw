package mailer

import (
	"os"
	"testing"
)

func TestMailtrap_Send(t *testing.T) {
	if len(os.Getenv("TEST_LIVE_REQUESTS")) == 0 {
		t.Skip("Not running LIVE REQUEST tests")
		return
	}

	user := os.Getenv("TEST_MAILTRAP_USER")
	pass := os.Getenv("TEST_MAILTRAP_PASSWORD")
	m, err := NewMailtrapMailer(MailtrapConfig{
		Server:   "smtp.mailtrap.io",
		Port:     587,
		User:     user,
		Password: pass,
	})

	if err != nil {
		t.Errorf("Failed to create Mailtrap mailer")
	}

	err = m.Send(Email{
		To: User{
			Name:    "FirstName LastName",
			Address: "test@gmail.com",
		},
		From: User{
			Name:    "FirstName LastName",
			Address: "test@example.com",
		},
		Subject: "Mailtrap Test Email",
		HTML:    "<html><hr> <b>BOLD TEXT</b>, non <i>bold</i></html>",
		Text:    "bold text, non bold",
	})

	if err != nil {
		t.Errorf("Failed to send email through Mailtrap %s", err)
		return
	}
}
