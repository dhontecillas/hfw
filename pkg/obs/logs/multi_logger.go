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

// Debug logs a message with the debug level
func (l *MultiLogger) Debug(msg string, attrMap map[string]interface{}) {
	for _, w := range l.wrapped {
		w.Debug(msg, attrMap)
	}
}

// Info logs a message with the info level
func (l *MultiLogger) Info(msg string, attrMap map[string]interface{}) {
	for _, w := range l.wrapped {
		w.Info(msg, attrMap)
	}
}

// Warn logs a message with the warn level
func (l *MultiLogger) Warn(msg string, attrMap map[string]interface{}) {
	for _, w := range l.wrapped {
		w.Warn(msg, attrMap)
	}
}

// Err logs a message with the error level
func (l *MultiLogger) Err(err error, msg string, attrMap map[string]interface{}) {
	for _, w := range l.wrapped {
		w.Err(err, msg, attrMap)
	}
}

// Fatal logs a message with the fatal level
func (l *MultiLogger) Fatal(msg string, attrMap map[string]interface{}) {
	for _, w := range l.wrapped {
		w.Fatal(msg, attrMap)
	}
}

// Panic logs a message with the fatal level
func (l *MultiLogger) Panic(msg string, attrMap map[string]interface{}) {
	for _, w := range l.wrapped {
		// here, probably the first logger to panic will
		// halt the program
		w.Panic(msg, attrMap)
	}
}

// Str adds a tag to the logger of type string
func (l *MultiLogger) Str(key, val string) {
	for _, w := range l.wrapped {
		w.Str(key, val)
	}
}

// I64 adds a tag to the logger of type int64
func (l *MultiLogger) I64(key string, val int64) {
	for _, w := range l.wrapped {
		w.I64(key, val)
	}
}

// F64 adds a tag to the logger of type float64
func (l *MultiLogger) F64(key string, val float64) {
	for _, w := range l.wrapped {
		w.F64(key, val)
	}
}

// Bool adds a tag to the logger of type bool
func (l *MultiLogger) Bool(key string, val bool) {
	for _, w := range l.wrapped {
		w.Bool(key, val)
	}
}

// Labels sets labels for a logger in a batch
func (l *MultiLogger) SetAttrs(attrMap map[string]interface{}) {
	for _, w := range l.wrapped {
		w.SetAttrs(attrMap)
	}
}
