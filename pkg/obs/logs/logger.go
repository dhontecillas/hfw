package logs

// Logger defines the interface to log data.
// Most of the calls return the same Logger interface so
// we can easily chain calls, like l.Str("foo", "bar").I64("count", 3).Trace("blab").
type Logger interface {
	// Trace logs a message with the trace level
	Trace(msg string)
	// Debug logs a message with the debug level
	Debug(msg string)
	// Info logs a message with the info level
	Info(msg string)
	// Warn logs a message with the warn level
	Warn(msg string)
	// Err logs a message with the error level
	Err(err error, msg string)
	// Fatal logs a message with the fatal level
	Fatal(msg string)
	// Panic logs a message with the fatal level
	Panic(msg string)

	// TraceMsg creates a LogMsg that inherit the logger
	// tags with a trace level.
	TraceMsg(msg string) LogMsg
	// DebugMsg creates a LogMsg that inherit the logger
	// tags with a debug level.
	DebugMsg(msg string) LogMsg
	// InfoMsg creates a LogMsg that inherit the logger
	// tags with a info level.
	InfoMsg(msg string) LogMsg
	// WarnMsg creates a LogMsg that inherit the logger
	// tags with a warn level.
	WarnMsg(msg string) LogMsg
	// ErrMsg creates a LogMsg that inherit the logger
	// tags with a error level.
	ErrMsg(err error, msg string) LogMsg
	// FatalMsg creates a LogMsg that inherit the logger
	// tags with a fatal level.
	FatalMsg(msg string) LogMsg
	// PanicMsg creates a LogMsg that inherit the logger
	// tags with a panic level.
	PanicMsg(msg string) LogMsg

	// Str adds a tag to the logger of type string
	Str(key, val string) Logger
	// I64 adds a tag to the logger of type int64
	I64(key string, val int64) Logger
	// F64 adds a tag to the logger of type float64
	F64(key string, val float64) Logger
	// Bool adds a tag to the logger of type bool
	Bool(key string, val bool) Logger

	// Labels sets labels for a logger in a batchl
	Labels(map[string]interface{}) Logger

	// Clone a logger, so from that point labels
	// will not be shared anymore.
	Clone() Logger
}

// LogMsg is a message entry that can be built and
// is not sent until is explicitly told to be sent.
type LogMsg interface {
	// Str adds a tag to the message of type string.
	Str(key, val string) LogMsg
	// I64 adds a tag to the message of type int64.
	I64(key string, val int64) LogMsg
	// F64 adds a tag to the message of type float64.
	F64(key string, val float64) LogMsg
	// Bool adds a tag to the message of type bool.
	Bool(key string, val bool) LogMsg

	// Send the message that has been constructed.
	Send()
}

// LoggerBuilderFn is the type required to instantiate a new Logger.
type LoggerBuilderFn func() Logger
