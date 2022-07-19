package mailer

import (
	"testing"
)

func TestMailer_NopMailer(t *testing.T) {
	m := NewNopMailer()
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
}
