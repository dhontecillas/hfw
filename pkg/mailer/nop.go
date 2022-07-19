package mailer

// NopMailer implements the Mailer interface, but does nothing when
// Send is called
type NopMailer struct{}

// NewNopMailer returns a Mailer that does absolutely nothing
func NewNopMailer() *NopMailer {
	return &NopMailer{}
}

// Send makes you believe that is sending an email, but it is not
func (m *NopMailer) Send(e Email) error {
	return nil
}
