package logs

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

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
	l.Level = logrus.DebugLevel
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

func (l *Logrus) entry() *logrus.Entry {
	c := GetCaller()
	l.dataMux.Lock()
	e := l.log.WithFields(l.data)
	l.dataMux.Unlock()
	e = e.WithTime(time.Now())
	e = e.WithField("file", fmt.Sprintf("%s:%d", c.File, c.Line))
	return e
}

// Clone clones a Logrus logger.
func (l *Logrus) Clone() Logger {
	l.dataMux.RLock()
	dataCap := len(l.data) + logrusCloneExtraDataCapacity
	if dataCap < logrusInitialDataCapacity {
		dataCap = logrusInitialDataCapacity
	}
	d := make(map[string]interface{}, dataCap)
	for k, v := range l.data {
		d[k] = v
	}
	l.dataMux.RUnlock()
	return &Logrus{
		data: d,
		log:  l.log,
	}
}

// Debug logs a message with the debug level
func (l *Logrus) Debug(msg string, attrMap map[string]interface{}) {
	e := l.entry()
	if attrMap != nil && len(attrMap) > 0 {
		e = e.WithFields(attrMap)
	}
	e.Debug(msg)
}

// Info logs a message with the info level
func (l *Logrus) Info(msg string, attrMap map[string]interface{}) {
	e := l.entry()
	if attrMap != nil && len(attrMap) > 0 {
		e = e.WithFields(attrMap)
	}
	e.Info(msg)
}

// Warn logs a message with the warn level
func (l *Logrus) Warn(msg string, attrMap map[string]interface{}) {
	e := l.entry()
	if attrMap != nil && len(attrMap) > 0 {
		e = e.WithFields(attrMap)
	}
	e.Warn(msg)
}

// Err logs a message with the error level
func (l *Logrus) Err(err error, msg string, attrMap map[string]interface{}) {
	e := l.entry()
	e = e.WithError(err)
	if attrMap != nil && len(attrMap) > 0 {
		e = e.WithFields(attrMap)
	}
	e.Error(err, msg)
}

// Fatal logs a message with the fatal level
func (l *Logrus) Fatal(msg string, attrMap map[string]interface{}) {
	e := l.entry()
	if attrMap != nil && len(attrMap) > 0 {
		e = e.WithFields(attrMap)
	}
	e.Fatal(msg)
}

// Panic logs a message with the fatal level
func (l *Logrus) Panic(msg string, attrMap map[string]interface{}) {
	e := l.entry()
	if attrMap != nil && len(attrMap) > 0 {
		e = e.WithFields(attrMap)
	}
	e.Panic(msg)
}

// Str adds a tag to the logger of type string
func (l *Logrus) Str(key, val string) {
	l.setAttr(key, val)
}

// I64 adds a tag to the logger of type int64
func (l *Logrus) I64(key string, val int64) {
	l.setAttr(key, val)
}

// F64 adds a tag to the logger of type float64
func (l *Logrus) F64(key string, val float64) {
	l.setAttr(key, val)
}

// Bool adds a tag to the logger of type bool
func (l *Logrus) Bool(key string, val bool) {
	l.setAttr(key, val)
}

func (l *Logrus) setAttr(key string, val interface{}) {
	l.dataMux.Lock()
	l.data[key] = val
	l.dataMux.Unlock()
}

// SetAttrs sets labels for a logger in a batch
func (l *Logrus) SetAttrs(attrMap map[string]interface{}) {
	l.dataMux.Lock()
	for k, v := range attrMap {
		l.data[k] = v
	}
	defer l.dataMux.Unlock()
}
