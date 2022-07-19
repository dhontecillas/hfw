package logs

// NopLogger is a logger that logs nothing.
type NopLogger struct {
}

// NopLoggerMsg is a log message that will send nothing.
type NopLoggerMsg struct {
}

// NewNopLogger is a function to construct a NopLogger logger.
func NewNopLogger() *NopLogger {
	return &NopLogger{}
}

// NewNopLoggerBuilder returns a function to construct NopLogger.
func NewNopLoggerBuilder() LoggerBuilderFn {
	return func() Logger {
		return NewNopLogger()
	}
}

// Clone clones a NopLogger logger.
func (nl *NopLogger) Clone() Logger {
	return &NopLogger{}
}

// Trace logs a message with the trace level
func (nl *NopLogger) Trace(msg string) {
}

// Debug logs a message with the debug level
func (nl *NopLogger) Debug(msg string) {
}

// Info logs a message with the info level
func (nl *NopLogger) Info(msg string) {
}

// Warn logs a message with the warn level
func (nl *NopLogger) Warn(msg string) {
}

// Err logs a message with the error level
func (nl *NopLogger) Err(err error, msg string) {
}

// Fatal logs a message with the fatal level
func (nl *NopLogger) Fatal(msg string) {
}

// Panic logs a message with the fatal level
func (nl *NopLogger) Panic(msg string) {
}

// Str adds a tag to the logger of type string
func (nl *NopLogger) Str(key, val string) Logger {
	return nl
}

// I64 adds a tag to the logger of type int64
func (nl *NopLogger) I64(key string, val int64) Logger {
	return nl
}

// F64 adds a tag to the logger of type float64
func (nl *NopLogger) F64(key string, val float64) Logger {
	return nl
}

// Bool adds a tag to the logger of type bool
func (nl *NopLogger) Bool(key string, val bool) Logger {
	return nl
}

// Labels sets labels for a logger in a batchl
func (nl *NopLogger) Labels(labels map[string]interface{}) Logger {
	return nl
}

// TraceMsg creates a LogMsg that inherit the logger
// tags with a trace level.
func (nl *NopLogger) TraceMsg(msg string) LogMsg {
	return &NopLoggerMsg{}
}

// DebugMsg creates a LogMsg that inherit the logger
// tags with a debug level.
func (nl *NopLogger) DebugMsg(msg string) LogMsg {
	return &NopLoggerMsg{}
}

// InfoMsg creates a LogMsg that inherit the logger
// tags with a info level.
func (nl *NopLogger) InfoMsg(msg string) LogMsg {
	return &NopLoggerMsg{}
}

// WarnMsg creates a LogMsg that inherit the logger
// tags with a warn level.
func (nl *NopLogger) WarnMsg(msg string) LogMsg {
	return &NopLoggerMsg{}
}

// ErrMsg creates a LogMsg that inherit the logger
// tags with a error level.
func (nl *NopLogger) ErrMsg(err error, msg string) LogMsg {
	return &NopLoggerMsg{}
}

// FatalMsg creates a LogMsg that inherit the logger
// tags with a fatal level.
func (nl *NopLogger) FatalMsg(msg string) LogMsg {
	return &NopLoggerMsg{}
}

// PanicMsg creates a LogMsg that inherit the logger
// tags with a panic level.
func (nl *NopLogger) PanicMsg(msg string) LogMsg {
	return &NopLoggerMsg{}
}

// Send the message that has been constructed.
func (lm *NopLoggerMsg) Send() {
}

// Str adds a tag to the message of type string.
func (lm *NopLoggerMsg) Str(key, val string) LogMsg {
	return lm
}

// I64 adds a tag to the message of type int64.
func (lm *NopLoggerMsg) I64(key string, val int64) LogMsg {
	return lm
}

// F64 adds a tag to the message of type float64.
func (lm *NopLoggerMsg) F64(key string, val float64) LogMsg {
	return lm
}

// Bool adds a tag to the message of type bool.
func (lm *NopLoggerMsg) Bool(key string, val bool) LogMsg {
	return lm
}
