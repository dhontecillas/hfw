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

// Debug logs a message with the debug level
func (nl *NopLogger) Debug(msg string, attrMap map[string]interface{}) {}

// Info logs a message with the info level
func (nl *NopLogger) Info(msg string, attrMap map[string]interface{}) {}

// Warn logs a message with the warn level
func (nl *NopLogger) Warn(msg string, attrMap map[string]interface{}) {}

// Err logs a message with the error level
func (nl *NopLogger) Err(err error, msg string, attrMap map[string]interface{}) {}

// Fatal logs a message with the fatal level
func (nl *NopLogger) Fatal(msg string, attrMap map[string]interface{}) {}

// Panic logs a message with the fatal level
func (nl *NopLogger) Panic(msg string, attrMap map[string]interface{}) {}

// Str adds a tag to the logger of type string
func (nl *NopLogger) Str(key, val string) {}

// I64 adds a tag to the logger of type int64
func (nl *NopLogger) I64(key string, val int64) {}

// F64 adds a tag to the logger of type float64
func (nl *NopLogger) F64(key string, val float64) {}

// Bool adds a tag to the logger of type bool
func (nl *NopLogger) Bool(key string, val bool) {}

// Labels sets labels for a logger in a batch
func (nl *NopLogger) SetAttrs(attrMap map[string]interface{}) {}
