package logs

import (
	"fmt"
)

// MultiLogger is a logger that can wrap several
// other loggers an send the logs those multiple
// other instances.
type MultiLogger struct {
	wrapped []Logger
}

// MultiLogMsg is a LogMsg that sends log mesages
// to multiple other loggers.
type MultiLogMsg struct {
	wrapped []LogMsg
}

// NewMultiLogger crates a new logger that will forward
// log messages to the list of wrapped loggers.
func NewMultiLogger(wrapped ...Logger) *MultiLogger {
	m := &MultiLogger{}
	m.wrapped = make([]Logger, len(wrapped))
	copy(m.wrapped, wrapped)
	return m
}

// NewMultiLoggerBuilder returns a function to create MultiLogger loggers.
func NewMultiLoggerBuilder(wrappedFns ...LoggerBuilderFn) LoggerBuilderFn {
	return func() Logger {
		m := &MultiLogger{}
		m.wrapped = make([]Logger, 0, len(wrappedFns))
		for idx, fn := range wrappedFns {
			if fn == nil {
				fmt.Printf("func %d is nil!\n", idx)
				continue
			}
			l := fn()
			if l != nil {
				m.wrapped = append(m.wrapped, l)
			}
		}
		return m
	}
}

// Clone a MultiLogger logger.
func (l *MultiLogger) Clone() Logger {
	lc := &MultiLogger{}
	lc.wrapped = make([]Logger, 0, len(l.wrapped))
	for _, w := range l.wrapped {
		lc.wrapped = append(lc.wrapped, w.Clone())
	}
	return lc
}

// Trace logs a message with the trace level
func (l *MultiLogger) Trace(msg string) {
	for _, w := range l.wrapped {
		w.Trace(msg)
	}
}

// Debug logs a message with the debug level
func (l *MultiLogger) Debug(msg string) {
	for _, w := range l.wrapped {
		w.Debug(msg)
	}
}

// Info logs a message with the info level
func (l *MultiLogger) Info(msg string) {
	for _, w := range l.wrapped {
		w.Info(msg)
	}
}

// Warn logs a message with the warn level
func (l *MultiLogger) Warn(msg string) {
	for _, w := range l.wrapped {
		w.Warn(msg)
	}
}

// Err logs a message with the error level
func (l *MultiLogger) Err(err error, msg string) {
	for _, w := range l.wrapped {
		w.Err(err, msg)
	}
}

// Fatal logs a message with the fatal level
func (l *MultiLogger) Fatal(msg string) {
	for _, w := range l.wrapped {
		w.Fatal(msg)
	}
}

// Panic logs a message with the fatal level
func (l *MultiLogger) Panic(msg string) {
	for _, w := range l.wrapped {
		// here, probably the first logger to panic will
		// halt the program
		w.Panic(msg)
	}
}

// TraceMsg creates a LogMsg that inherit the logger
// tags with a trace level.
func (l *MultiLogger) TraceMsg(msg string) LogMsg {
	lm := &MultiLogMsg{
		wrapped: make([]LogMsg, 0, len(l.wrapped)),
	}
	for _, w := range l.wrapped {
		lm.wrapped = append(lm.wrapped, w.TraceMsg(msg))
	}
	return lm
}

// DebugMsg creates a LogMsg that inherit the logger
// tags with a debug level.
func (l *MultiLogger) DebugMsg(msg string) LogMsg {
	lm := &MultiLogMsg{
		wrapped: make([]LogMsg, 0, len(l.wrapped)),
	}
	for _, w := range l.wrapped {
		lm.wrapped = append(lm.wrapped, w.DebugMsg(msg))
	}
	return lm
}

// InfoMsg creates a LogMsg that inherit the logger
// tags with a info level.
func (l *MultiLogger) InfoMsg(msg string) LogMsg {
	lm := &MultiLogMsg{
		wrapped: make([]LogMsg, 0, len(l.wrapped)),
	}
	for _, w := range l.wrapped {
		lm.wrapped = append(lm.wrapped, w.InfoMsg(msg))
	}
	return lm
}

// WarnMsg creates a LogMsg that inherit the logger
// tags with a warn level.
func (l *MultiLogger) WarnMsg(msg string) LogMsg {
	lm := &MultiLogMsg{
		wrapped: make([]LogMsg, 0, len(l.wrapped)),
	}
	for _, w := range l.wrapped {
		lm.wrapped = append(lm.wrapped, w.WarnMsg(msg))
	}
	return lm
}

// ErrMsg creates a LogMsg that inherit the logger
// tags with a error level.
func (l *MultiLogger) ErrMsg(err error, msg string) LogMsg {
	lm := &MultiLogMsg{
		wrapped: make([]LogMsg, 0, len(l.wrapped)),
	}
	for _, w := range l.wrapped {
		lm.wrapped = append(lm.wrapped, w.ErrMsg(err, msg))
	}
	return lm
}

// FatalMsg creates a LogMsg that inherit the logger
// tags with a fatal level.
func (l *MultiLogger) FatalMsg(msg string) LogMsg {
	lm := &MultiLogMsg{
		wrapped: make([]LogMsg, 0, len(l.wrapped)),
	}
	for _, w := range l.wrapped {
		lm.wrapped = append(lm.wrapped, w.FatalMsg(msg))
	}
	return lm
}

// PanicMsg creates a LogMsg that inherit the logger
// tags with a panic level.
func (l *MultiLogger) PanicMsg(msg string) LogMsg {
	lm := &MultiLogMsg{
		wrapped: make([]LogMsg, 0, len(l.wrapped)),
	}
	for _, w := range l.wrapped {
		lm.wrapped = append(lm.wrapped, w.PanicMsg(msg))
	}
	return lm
}

// Str adds a tag to the logger of type string
func (l *MultiLogger) Str(key, val string) Logger {
	for _, w := range l.wrapped {
		w.Str(key, val)
	}
	return l
}

// I64 adds a tag to the logger of type int64
func (l *MultiLogger) I64(key string, val int64) Logger {
	for _, w := range l.wrapped {
		w.I64(key, val)
	}
	return l
}

// F64 adds a tag to the logger of type float64
func (l *MultiLogger) F64(key string, val float64) Logger {
	for _, w := range l.wrapped {
		w.F64(key, val)
	}
	return l
}

// Bool adds a tag to the logger of type bool
func (l *MultiLogger) Bool(key string, val bool) Logger {
	for _, w := range l.wrapped {
		w.Bool(key, val)
	}
	return l
}

// Labels sets labels for a logger in a batch
func (l *MultiLogger) Labels(labels map[string]interface{}) Logger {
	for _, w := range l.wrapped {
		w.Labels(labels)
	}
	return l
}

// Send the message that has been constructed.
func (lm *MultiLogMsg) Send() {
	for _, w := range lm.wrapped {
		w.Send()
	}
}

// Str adds a tag to the message of type string.
func (lm *MultiLogMsg) Str(key, val string) LogMsg {
	for _, w := range lm.wrapped {
		w.Str(key, val)
	}
	return lm
}

// I64 adds a tag to the message of type int64.
func (lm *MultiLogMsg) I64(key string, val int64) LogMsg {
	for _, w := range lm.wrapped {
		w.I64(key, val)
	}
	return lm
}

// F64 adds a tag to the message of type float64.
func (lm *MultiLogMsg) F64(key string, val float64) LogMsg {
	for _, w := range lm.wrapped {
		w.F64(key, val)
	}
	return lm
}

// Bool adds a tag to the message of type bool.
func (lm *MultiLogMsg) Bool(key string, val bool) LogMsg {
	for _, w := range lm.wrapped {
		w.Bool(key, val)
	}
	return lm
}
