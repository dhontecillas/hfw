package extdeps

import (
	"github.com/dhontecillas/hfw/pkg/db"
	"github.com/dhontecillas/hfw/pkg/mailer"
	"github.com/dhontecillas/hfw/pkg/notifications"
	"github.com/dhontecillas/hfw/pkg/obs"
)

// ExternalServicesBuilder holds references to services
// that could be needed to perform any operation
// and that must be configured at startup time, like
// and email sender, a sql db ...
type ExternalServicesBuilder struct {
	MailSender mailer.Mailer
	SQL        db.SQLDB
	Composer   notifications.Composer

	// global configured insigher from where we will clone
	insBuilder obs.InsighterBuilderFn
	insFlush   func()
}

// NewExternalServicesBuilder creates a new ExternalServices instance
func NewExternalServicesBuilder(
	insighterBuilderFn obs.InsighterBuilderFn,
	insighterFlushFn func(),
	mailSender mailer.Mailer,
	sql db.SQLDB,
	composer notifications.Composer) *ExternalServicesBuilder {

	return &ExternalServicesBuilder{
		MailSender: mailSender,
		SQL:        sql,
		Composer:   composer,
		insBuilder: insighterBuilderFn,
		insFlush:   insighterFlushFn,
	}
}

// Shutdown all the service references inside
// the ExternalServices
func (es *ExternalServicesBuilder) Shutdown() {
	es.SQL.Close()
	if es.insFlush != nil {
		es.insFlush()
	}
}

// Insighter returns an Insighter instance
func (es *ExternalServicesBuilder) Insighter() *obs.Insighter {
	return es.insBuilder()
}

// ExtServices returns a new ExtServices to be used
func (es *ExternalServicesBuilder) ExtServices() *ExtServices {
	return &ExtServices{
		MailSender: es.MailSender,
		SQL:        es.SQL,
		Composer:   es.Composer,
		Ins:        es.insBuilder(),
	}
}
