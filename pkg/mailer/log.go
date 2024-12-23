package mailer

import (
	"time"

	"github.com/dhontecillas/hfw/pkg/obs"
)

// LoggerMailer implements a Mailer interface, wrapping another Mailer, and
// adding log information to the Send method
type LoggerMailer struct {
	wrapped Mailer
	ins     *obs.Insighter
}

// NewLoggerMailer wraps a given mailer to add logging information
func NewLoggerMailer(wrapped Mailer, ins *obs.Insighter) *LoggerMailer {
	return &LoggerMailer{
		wrapped: wrapped,
		ins:     ins,
	}
}

// Send sends an email through the wrapped mailer and logs information
// about the email being sent
func (m *LoggerMailer) Send(e Email) error {
	startTime := time.Now()
	err := m.wrapped.Send(e)
	elapsed := time.Since(startTime)
	if err != nil {
		m.ins.L.Err(err, "mailer send error", map[string]interface{}{
			"started": startTime.String(),
			"elapsed": elapsed.String(),
			"error":   err.Error(),
		})
	} else {
		m.ins.L.Info("mailer send email", map[string]interface{}{
			"started": startTime.String(),
			"elapsed": elapsed.String(),
		})
	}
	return err
}

// Sender proxies the call to the underlying mailer
func (m *LoggerMailer) Sender() (string, string) {
	return m.wrapped.Sender()
}
