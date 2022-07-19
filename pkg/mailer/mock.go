package mailer

// MockMailer implements the Mailer interface, saving emails to be
// sent in memory (SentMails) instead of actually sending them
type MockMailer struct {
	SentMails []Email
}

// NewMockMailer creates a mailre that stores in memory all sent emails
// WARNING: Not for production!! Use only for tests
func NewMockMailer() *MockMailer {
	return &MockMailer{
		SentMails: make([]Email, 0),
	}
}

// Send stores the email to be sent in the SentMails field
func (m *MockMailer) Send(e Email) error {
	m.SentMails = append(m.SentMails, Email{
		To:      e.To,
		From:    e.From,
		Subject: e.Subject,
		HTML:    e.HTML,
		Text:    e.Text,
	})
	return nil
}
