package mailer

import "fmt"

// ConsoleMailer implements the Mailer interface to print summaries
// of emails insted of actually send them.
type ConsoleMailer struct{}

// NewConsoleMailer returns a new Mailer that prints emails instead of
// sending a real email
func NewConsoleMailer() *ConsoleMailer {
	return &ConsoleMailer{}
}

// Send prints the email to, from and subject fields to console
func (m *ConsoleMailer) Send(e Email) error {
	fmt.Printf("Email SENT: %s\n", e)
	return nil
}

func (m *ConsoleMailer) Sender() (string, string) {
	return "noreply@example.com", "No Reply"
}
