package extdeps

import (
	"github.com/dhontecillas/hfw/pkg/db"
	"github.com/dhontecillas/hfw/pkg/mailer"
	"github.com/dhontecillas/hfw/pkg/notifications"
	"github.com/dhontecillas/hfw/pkg/obs"
)

// ExternalServicesBuilderFn defines the signature
// for building a ExternalServices.
// Usually used to create new instances for each served
// request
type ExternalServicesBuilderFn func() *ExternalServices

// ExtServices holds thread-safe instances to make use of
// external services.
type ExtServices struct {
	MailSender mailer.Mailer
	SQL        db.SQLDB
	Notifier   notifications.Notifier
	Ins        *obs.Insighter
}

// Clone creates a shallow copy of the external services,
// except for the insighter one, that will execute a clone.
func (ed *ExtServices) Clone() *ExtServices {
	return &ExtServices{
		MailSender: ed.MailSender,
		SQL:        ed.SQL,
		Notifier:   ed.Notifier,
		Ins:        ed.Ins.Clone(),
	}
}
