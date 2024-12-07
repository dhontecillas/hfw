package extdeps

import (
	"github.com/dhontecillas/hfw/pkg/mailer"
	"github.com/dhontecillas/hfw/pkg/notifications"
	"github.com/dhontecillas/hfw/pkg/obs"
	"github.com/dhontecillas/hfw/pkg/obs/logs"
	"github.com/dhontecillas/hfw/pkg/obs/metrics"
	"github.com/dhontecillas/hfw/pkg/obs/traces"
)

// GetNopExternalServices creates No Op services:
// a mailer that does not sends emails, a notifier
// that does not send notifications, and an insighter
// that does not log, send metrics or traces.
// Useful for testing.
func GetNopExternalServices() *ExternalServicesBuilder {
	logBuilder, _, _ := logs.NewLogrusBuilder(nil)
	nopMeterBuilder, _ := metrics.NewNopMeterBuilder()
	nopTracerBuilder := traces.NewNopTracerBuilder()

	insBuilder := obs.NewInsighterBuilder(logBuilder, nopMeterBuilder, nopTracerBuilder)

	// TODO: we need a way to porovide a No-Op SQLDB interface, that Master returns nil !
	return NewExternalServicesBuilder(insBuilder, func() {}, mailer.NewNopMailer(),
		nil, notifications.NewNopComposer())
}
