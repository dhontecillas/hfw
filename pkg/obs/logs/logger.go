package logs

import (
	"github.com/dhontecillas/hfw/pkg/obs/attrs"
)

// Logger defines the interface to log data.
// Most of the calls return the same Logger interface so
// we can easily chain calls, like l.Str("foo", "bar").I64("count", 3).Trace("blab").
type Logger interface {
	attrs.Attributable

	// Debug logs a message with the debug level
	// the attributes passed as parameters are used only
	// for the sent meessage.
	Debug(msg string, attrs map[string]interface{})
	// Info logs a message with the info level
	// the attributes passed as parameters are used only
	// for the sent meessage.
	Info(msg string, attrs map[string]interface{})
	// Err logs a message with the error level
	// the attributes passed as parameters are used only
	// for the sent meessage.
	Err(err error, msg string, attrs map[string]interface{})
	// Warn logs a message with the warn level
	// the attributes passed as parameters are used only
	// for the sent meessage.
	Warn(msg string, attrs map[string]interface{})
	// Fatal logs a message with the fatal level
	// the attributes passed as parameters are used only
	// for the sent meessage.
	Fatal(msg string, attrs map[string]interface{})
	// Panic logs a message with the panic level
	// the attributes passed as parameters are used only
	// for the sent meessage.
	Panic(msg string, attrs map[string]interface{})

	// Clone a logger, so from that attributes
	// will not be shared anymore.
	Clone() Logger
}

// LoggerBuilderFn is the type required to instantiate a new Logger.
type LoggerBuilderFn func() Logger
