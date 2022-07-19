package extdeps

import (
	"github.com/dhontecillas/hfw/pkg/db"
	"github.com/dhontecillas/hfw/pkg/mailer"
	"github.com/dhontecillas/hfw/pkg/notifications"
	"github.com/dhontecillas/hfw/pkg/obs"
)

// ExternalServices holds references to services
// that could be needed to perform any operation
// and that must be configured at startup time, like
// and email sender, a sql db ...
type ExternalServices struct {
	MailSender mailer.Mailer
	SQL        db.SQLDB
	Notifier   notifications.Notifier

	// global configured insigher from where we will clone
	ins      *obs.Insighter
	insFlush func()
}

// NewExternalServices creates a new ExternalServices instance
func NewExternalServices(
	insighterBuilderFn obs.InsighterBuilderFn,
	insighterFlushFn func(),
	mailSender mailer.Mailer,
	sql db.SQLDB,
	notifier notifications.Notifier) *ExternalServices {

	ins := insighterBuilderFn()
	return &ExternalServices{
		MailSender: mailSender,
		SQL:        sql,
		Notifier:   notifier,
		ins:        ins,
		insFlush:   insighterFlushFn,
	}
}

// Shutdown all the service references inside
// the ExternalServices
func (es *ExternalServices) Shutdown() {
	es.SQL.Close()
	if es.insFlush != nil {
		es.insFlush()
	}
}

// Insighter returns an Insighter instance
func (es *ExternalServices) Insighter() *obs.Insighter {
	return es.ins
}

// ExtServices returns a new ExtServices to be used
func (es *ExternalServices) ExtServices() *ExtServices {
	return &ExtServices{
		MailSender: es.MailSender,
		SQL:        es.SQL,
		Notifier:   es.Notifier,
		Ins:        es.ins.Clone(),
	}
}
