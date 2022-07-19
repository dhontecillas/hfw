package mailer

import (
	"testing"
)

func TestMailer_MockMailer(t *testing.T) {
	m := NewMockMailer()

	res := m.Send(Email{
		To:      User{Name: "foo", Address: "foo@example.com"},
		From:    User{Name: "bar", Address: "bar@example.com"},
		Subject: "FooBar",
		HTML:    "<html><b>Body</b></html>",
		Text:    "Body",
	})
	if res != nil {
		t.Errorf("Unexpected error %s", res)
		return
	}
	if len(m.SentMails) != 1 {
		t.Errorf("Wanted 1 email, Got %d", len(m.SentMails))
		return
	}
	e := m.SentMails[0]
	if e.To.String() != "\"foo\" <foo@example.com>" {
		t.Errorf("Mail To fail in mock is not recorded")
		return
	}
}
