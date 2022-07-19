package mailer

import (
	"testing"

	"github.com/dhontecillas/hfw/pkg/obs"
	"github.com/dhontecillas/hfw/pkg/obs/logs"
	"github.com/dhontecillas/hfw/pkg/obs/metrics"
	"github.com/dhontecillas/hfw/pkg/obs/traces"
)

func TestMailer_LogWrapper(t *testing.T) {
	logFn := logs.NewNopLoggerBuilder()
	meterFn, _ := metrics.NewNopMeterBuilder()
	tracerFn := traces.NewNopTracerBuilder()
	insBuilder := obs.NewInsighterBuilder([]obs.TagDefinition{},
		logFn, meterFn, tracerFn)
	ins := insBuilder()

	mm := NewMockMailer()
	lm := NewLoggerMailer(mm, ins)

	res := lm.Send(Email{
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

	if len(mm.SentMails) != 1 {
		t.Errorf("Wanted 1 email, Got %d", len(mm.SentMails))
		return
	}

	e := mm.SentMails[0]
	if e.To.String() != "\"foo\" <foo@example.com>" {
		t.Errorf("Mail To fail in mock is not recorded")
		return
	}
}
