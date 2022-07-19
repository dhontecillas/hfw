package mailer

import (
	"fmt"
	"strings"
)

// User is the struct used for defining sender recipient info
type User struct {
	Name    string
	Address string
}

// Email contains all the required information for an email
type Email struct {
	To      User
	From    User
	Subject string
	HTML    string
	Text    string
}

// String implements the Stringer interface to print user addresses
func (u User) String() string {
	return fmt.Sprintf("\"%s\" <%s>", u.Name, u.Address)
}

func summarizeContent(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	// remove all new lines
	toRemove := []string{"\r", "\n"}
	for _, ch := range toRemove {
		s = strings.ReplaceAll(s, ch, "")
	}
	s = strings.Trim(s, "\r\n ")
	if len(s) <= maxLen {
		return s
	}
	// if there are more bytes than maxLen, the cut it at the
	// correct rune index:
	cnt := 0
	for idx := range s {
		if cnt >= maxLen {
			return s[:idx]
		}
		cnt++
	}
	return s
}

// String implements the Stringer interface to print email summary
func (e Email) String() string {
	return fmt.Sprintf("To: %s, From: %s, Subject: %s, Text: %s", e.To, e.From, e.Subject,
		summarizeContent(e.Text, 1024))
}

// Mailer is the general interface for sending emails
type Mailer interface {
	Send(e Email) error
}
