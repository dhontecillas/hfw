package logs

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
)

const logrusInitialDataCapacity = 16
const logrusCloneExtraDataCapacity = 4

// LogrusConf contains the values to configure
// a Logrus parser.
type LogrusConf struct {
	// OutFileName is the file were we want to output the logs.
	OutFileName string

	// GraylogHost if not empty will report directly to graylog.
	GraylogHost string
	// GraylogFieldPrefix to add a prefix to all graylog logs.
	GraylogFieldPrefix string
}

// Logrus is the logger to work with logrus.
type Logrus struct {
	data    map[string]interface{}
	dataMux sync.RWMutex
	log     *logrus.Logger
}

// NewLogrus creates a new Logrus logger
func NewLogrus(w io.Writer) *Logrus {
	l := logrus.New()
	if w == nil {
		l.Out = os.Stdout
	} else {
		l.Out = w
	}
	l.Level = logrus.TraceLevel
	l.SetFormatter(&logrus.JSONFormatter{})
	d := make(map[string]interface{}, logrusInitialDataCapacity)

	return &Logrus{
		data: d,
		log:  l,
	}
}

// NewLogrusBuilder returns a function to create Logrus loggers
func NewLogrusBuilder(conf *LogrusConf) (LoggerBuilderFn, func(), error) {
	if conf == nil {
		conf = &LogrusConf{}
	}
	flush := func() {}

	f := os.Stdout
	if len(conf.OutFileName) != 0 {
		var err error
		f, err = os.OpenFile(conf.OutFileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return nil, nil, err
		}
	}

	parentLog := NewLogrus(f)
	if len(conf.GraylogHost) > 0 {
		h := NewCustomGraylogHook(conf.GraylogHost, conf.GraylogFieldPrefix)
		parentLog.log.AddHook(h)
		flush = func() {
			h.Flush()
		}
	}

	return func() Logger {
		return parentLog.Clone()
	}, flush, nil
}

func (lr *Logrus) entry() *logrus.Entry {
	c := GetCaller()
	lr.dataMux.Lock()
	lr.data["file"] = fmt.Sprintf("%s:%d", c.File, c.Line)
	e := lr.log.WithFields(lr.data)
	lr.dataMux.Unlock()
	return e
}

// Clone clones a Logrus logger.
func (lr *Logrus) Clone() Logger {
	lr.dataMux.RLock()
	dataCap := len(lr.data) + logrusCloneExtraDataCapacity
	if dataCap < logrusInitialDataCapacity {
		dataCap = logrusInitialDataCapacity
	}
	d := make(map[string]interface{}, dataCap)
	for k, v := range lr.data {
		d[k] = v
	}
	lr.dataMux.RUnlock()
	return &Logrus{
		data: d,
		log:  lr.log,
	}
}

func (lr *Logrus) copyData() map[string]interface{} {
	lr.dataMux.RLock()
	dataCap := len(lr.data) + logrusCloneExtraDataCapacity
	if dataCap < logrusInitialDataCapacity {
		dataCap = logrusInitialDataCapacity
	}
	d := make(map[string]interface{}, dataCap)
	for k, v := range lr.data {
		d[k] = v
	}
	lr.dataMux.RUnlock()
	return d
}

// Trace logs a message with the trace level
func (lr *Logrus) Trace(msg string) {
	lr.entry().Trace(msg)
}

// Debug logs a message with the debug level
func (lr *Logrus) Debug(msg string) {
	lr.entry().Debug(msg)
}

// Info logs a message with the info level
func (lr *Logrus) Info(msg string) {
	lr.entry().Info(msg)
}

// Warn logs a message with the warn level
func (lr *Logrus) Warn(msg string) {
	lr.entry().Warn(msg)
}

// Err logs a message with the error level
func (lr *Logrus) Err(err error, msg string) {
	lr.entry().Error(err, msg)
}

// Fatal logs a message with the fatal level
func (lr *Logrus) Fatal(msg string) {
	lr.entry().Fatal(msg)
}

// Panic logs a message with the fatal level
func (lr *Logrus) Panic(msg string) {
	lr.entry().Panic(msg)
}

// TraceMsg creates a LogMsg that inherit the logger
// tags with a trace level.
func (lr *Logrus) TraceMsg(msg string) LogMsg {
	lm := &LogrusMsg{
		entry: lr.entry(),
		data:  lr.copyData(),
	}
	lm.entry.Level = logrus.TraceLevel
	lm.entry.Message = msg
	return lm
}

// DebugMsg creates a LogMsg that inherit the logger
// tags with a debug level.
func (lr *Logrus) DebugMsg(msg string) LogMsg {
	lm := &LogrusMsg{
		entry: lr.entry(),
		data:  lr.copyData(),
	}
	lm.entry.Level = logrus.DebugLevel
	lm.entry.Message = msg
	return lm
}

// InfoMsg creates a LogMsg that inherit the logger
// tags with a info level.
func (lr *Logrus) InfoMsg(msg string) LogMsg {
	lm := &LogrusMsg{
		entry: lr.entry(),
		data:  lr.copyData(),
	}
	lm.entry.Level = logrus.InfoLevel
	lm.entry.Message = msg
	return lm
}

// WarnMsg creates a LogMsg that inherit the logger
// tags with a warn level.
func (lr *Logrus) WarnMsg(msg string) LogMsg {
	lm := &LogrusMsg{
		entry: lr.entry(),
		data:  lr.copyData(),
	}
	lm.entry.Level = logrus.WarnLevel
	lm.entry.Message = msg
	return lm
}

// ErrMsg creates a LogMsg that inherit the logger
// tags with a error level.
func (lr *Logrus) ErrMsg(err error, msg string) LogMsg {
	lm := &LogrusMsg{
		entry: lr.entry(),
		data:  lr.copyData(),
	}
	lm.entry.Level = logrus.ErrorLevel
	lm.entry.Message = msg
	if err != nil {
		lm.data["error"] = err.Error()
	}
	return lm
}

// FatalMsg creates a LogMsg that inherit the logger
// tags with a fatal level.
func (lr *Logrus) FatalMsg(msg string) LogMsg {
	lm := &LogrusMsg{
		entry: lr.entry(),
		data:  lr.copyData(),
	}
	lm.entry.Level = logrus.FatalLevel
	lm.entry.Message = msg
	return lm
}

// PanicMsg creates a LogMsg that inherit the logger
// tags with a panic level.
func (lr *Logrus) PanicMsg(msg string) LogMsg {
	lm := &LogrusMsg{
		entry: lr.entry(),
		data:  lr.copyData(),
	}
	lm.entry.Level = logrus.PanicLevel
	lm.entry.Message = msg
	return lm
}

// Str adds a tag to the logger of type string
func (lr *Logrus) Str(key, val string) Logger {
	lr.dataMux.Lock()
	lr.data[key] = val
	lr.dataMux.Unlock()
	return lr
}

// I64 adds a tag to the logger of type int64
func (lr *Logrus) I64(key string, val int64) Logger {
	lr.dataMux.Lock()
	lr.data[key] = val
	lr.dataMux.Unlock()
	return lr
}

// F64 adds a tag to the logger of type float64
func (lr *Logrus) F64(key string, val float64) Logger {
	lr.dataMux.Lock()
	lr.data[key] = val
	lr.dataMux.Unlock()
	return lr
}

// Bool adds a tag to the logger of type bool
func (lr *Logrus) Bool(key string, val bool) Logger {
	lr.dataMux.Lock()
	lr.data[key] = val
	lr.dataMux.Unlock()
	return lr
}

// Labels sets labels for a logger in a batch
func (lr *Logrus) Labels(labels map[string]interface{}) Logger {
	lr.dataMux.Lock()
	defer lr.dataMux.Unlock()
	for k, v := range labels {
		lr.data[k] = v
	}
	return lr
}

// LogrusMsg is a message entry that can be built and
// is not sent until is explicitly told to be sent.
type LogrusMsg struct {
	entry *logrus.Entry
	data  map[string]interface{}
}

// Send the message that has been constructed.
func (lm *LogrusMsg) Send() {
	if lm.entry.Level == logrus.ErrorLevel {
		lm.entry.WithFields(lm.data).Log(lm.entry.Level, lm.entry.Message)
	} else {
		lm.entry.WithFields(lm.data).Log(lm.entry.Level, lm.entry.Message)
	}
}

// Str adds a tag to the message of type string.
func (lm *LogrusMsg) Str(key, val string) LogMsg {
	lm.data[key] = val
	return lm
}

// I64 adds a tag to the message of type int64.
func (lm *LogrusMsg) I64(key string, val int64) LogMsg {
	lm.data[key] = val
	return lm
}

// F64 adds a tag to the message of type float64.
func (lm *LogrusMsg) F64(key string, val float64) LogMsg {
	lm.data[key] = val
	return lm
}

// Bool adds a tag to the message of type bool.
func (lm *LogrusMsg) Bool(key string, val bool) LogMsg {
	lm.data[key] = val
	return lm
}
