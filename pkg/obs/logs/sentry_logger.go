package logs

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dhontecillas/hfw/pkg/obs/attrs"
	"github.com/getsentry/sentry-go"
)

var (
	sentryOnceInitialization  sync.Once
	sentryFlushFn             func()
	sentryParentHub           *sentry.Hub
	sentryInitializationError error
)

// SentryConf contains the values to initialize
// a Sentry logger.
type SentryConf struct {
	Dsn              string  `json:"dsn"`
	AttachStacktrace bool    `json:"attach_stacktrace"`
	SampleRate       float64 `json:"sample_rate"`
	Release          string  `json:"release"`
	Environment      string  `json:"environment"`
	FlushTimeoutSecs int     `json:"flush_timeout_secs`
	LevelThreshold   string  `json:"level_threshold"` // the mininimum level required to be sent
	AllowedTags      []string
}

// NewSentryConf creates a basic SentryConf
func NewSentryConf() *SentryConf {
	return &SentryConf{
		SampleRate: 1.0,
	}
}

// SentryLogger implements a logger that sends messages
// to Sentry.
type SentryLogger struct {
	hub         *sentry.Hub
	skipLevels  map[sentry.Level]bool
	allowedTags map[string]struct{}
}

// SentryLoggerMsg implements a LogMessage that can be sent
// to Sentry.
type SentryLoggerMsg struct {
	hub         *sentry.Hub
	entry       *sentry.Event
	err         error
	skipLevels  map[sentry.Level]bool
	allowedTags map[string]struct{}
}

// newSentryLogger returns a logger sending errors
// to sentry, and a flush function to send pending messages
// on a clean shutdown.
func newSentryLogger(conf *SentryConf) (*SentryLogger, func()) {
	sentryOnceInitialization.Do(func() {
		debug := false
		if strings.ToUpper(conf.Environment) == "DEBUG" {
			debug = true
		}

		err := sentry.Init(sentry.ClientOptions{
			Dsn:              conf.Dsn,
			AttachStacktrace: conf.AttachStacktrace,
			SampleRate:       conf.SampleRate,
			Debug:            debug,
			Environment:      conf.Environment,
		})

		if err != nil {
			sentryInitializationError = err
		}

		flushTimeout := time.Duration(conf.FlushTimeoutSecs) * time.Second
		sentryFlushFn = func() { sentry.Flush(flushTimeout) }
		sentryParentHub = sentry.CurrentHub()
	})

	disableLevels := map[sentry.Level]map[sentry.Level]bool{
		sentry.LevelInfo: map[sentry.Level]bool{
			sentry.LevelDebug: true,
		},
		sentry.LevelWarning: map[sentry.Level]bool{
			sentry.LevelDebug: true,
			sentry.LevelInfo:  true,
		},
		sentry.LevelError: map[sentry.Level]bool{
			sentry.LevelDebug: true,
			sentry.LevelInfo:  true,
		},
		sentry.LevelFatal: map[sentry.Level]bool{
			sentry.LevelDebug: true,
			sentry.LevelInfo:  true,
			sentry.LevelError: true,
		},
	}
	v, ok := disableLevels[sentry.Level(conf.LevelThreshold)]
	if !ok {
		v = map[sentry.Level]bool{}
	}

	allowedTags := make(map[string]struct{}, len(conf.AllowedTags))
	for _, tag := range conf.AllowedTags {
		allowedTags[tag] = struct{}{}
	}
	return &SentryLogger{
		hub:         sentryParentHub,
		skipLevels:  v,
		allowedTags: allowedTags,
	}, sentryFlushFn
}

// NewSentryBuilder returns a function to create Sentry loggers.
func NewSentryBuilder(conf *SentryConf) (LoggerBuilderFn, func(), error) {
	if conf == nil {
		conf = &SentryConf{}
	}
	masterSentryLogger, flushFn := newSentryLogger(conf)
	if sentryParentHub == nil {
		return nil, nil, sentryInitializationError
	}
	allowedTags := make(map[string]struct{}, len(conf.AllowedTags))
	for _, tag := range conf.AllowedTags {
		allowedTags[tag] = struct{}{}
	}
	return func() Logger {
		return &SentryLogger{
			hub:         sentryParentHub.Clone(),
			skipLevels:  masterSentryLogger.skipLevels,
			allowedTags: allowedTags,
		}
	}, flushFn, nil
}

func (l *SentryLogger) entry() *SentryLoggerMsg {
	c := GetCaller()
	file := fmt.Sprintf("%s:%d", c.File, c.Line)
	e := sentry.NewEvent()
	e.Extra["file"] = file
	e.Tags["filepos"] = file
	return &SentryLoggerMsg{
		hub:         l.hub,
		entry:       e,
		err:         nil,
		skipLevels:  l.skipLevels,
		allowedTags: l.allowedTags,
	}
}

// Clone clones a Sentry logger.
func (l *SentryLogger) Clone() Logger {
	return &SentryLogger{
		hub:         l.hub.Clone(),
		skipLevels:  l.skipLevels,
		allowedTags: l.allowedTags,
	}
}

// Debug logs a message with the debug level
func (l *SentryLogger) Debug(msg string, attrMap map[string]interface{}) {
	if l.skipLevels[sentry.LevelDebug] {
		return
	}
	e := l.entry()
	e.entry.Level = sentry.LevelDebug
	e.entry.Message = msg
	e.SetAttrs(attrMap)
	e.Send()
}

// Info logs a message with the info level
func (l *SentryLogger) Info(msg string, attrMap map[string]interface{}) {
	if l.skipLevels[sentry.LevelInfo] {
		return
	}
	e := l.entry()
	e.entry.Level = sentry.LevelInfo
	e.entry.Message = msg
	e.SetAttrs(attrMap)
	e.Send()
}

// Warn logs a message with the warn level
func (l *SentryLogger) Warn(msg string, attrMap map[string]interface{}) {
	if l.skipLevels[sentry.LevelWarning] {
		return
	}
	e := l.entry()
	e.entry.Level = sentry.LevelWarning
	e.entry.Message = msg
	e.SetAttrs(attrMap)
	e.Send()
}

// Err logs a message with the error level
func (l *SentryLogger) Err(err error, msg string, attrMap map[string]interface{}) {
	if l.skipLevels[sentry.LevelError] {
		return
	}
	// internally, CaptureException, unwraps all errors just sets the log level
	// and uses l.hub.CaptureException() , but it does not set the `.Message`
	// field. So, we put the message in the Extra fields
	e := l.entry()
	e.entry.Level = sentry.LevelError
	e.entry.Message = msg
	e.err = err
	e.SetAttrs(attrMap)
	e.Send()
}

// Fatal logs a message with the fatal level
func (l *SentryLogger) Fatal(msg string, attrMap map[string]interface{}) {
	e := l.entry()
	e.entry.Level = sentry.LevelFatal
	e.entry.Message = msg
	e.SetAttrs(attrMap)
	e.Send()
}

// Panic logs a message with the fatal level
func (l *SentryLogger) Panic(msg string, attrMap map[string]interface{}) {
	e := l.entry()
	e.entry.Level = sentry.LevelFatal // we do not have Panic level in sentry
	e.entry.Message = msg
	e.SetAttrs(attrMap)
	e.Send()
}

// Str adds a tag to the logger of type string
func (l *SentryLogger) Str(key, val string) {
	s := l.hub.Scope()
	if s == nil {
		return
	}
	if _, ok := l.allowedTags[key]; ok {
		s.SetTag(key, val)
	} else {
		s.SetExtra(key, val)
	}
}

// I64 adds a tag to the logger of type int64
func (l *SentryLogger) I64(key string, val int64) {
	s := l.hub.Scope()
	if s == nil {
		return
	}
	if _, ok := l.allowedTags[key]; ok {
		s.SetTag(key, strconv.FormatInt(val, 10))
	} else {
		s.SetExtra(key, val)
	}
}

// F64 adds a tag to the logger of type float64
func (l *SentryLogger) F64(key string, val float64) {
	s := l.hub.Scope()
	if s == nil {
		return
	}
	if _, ok := l.allowedTags[key]; ok {
		s.SetTag(key, strconv.FormatFloat(val, 'G', 9, 64))
	} else {
		s.SetExtra(key, val)
	}
}

// Bool adds a tag to the logger of type bool
func (l *SentryLogger) Bool(key string, val bool) {
	s := l.hub.Scope()
	if s == nil {
		return
	}
	if _, ok := l.allowedTags[key]; ok {
		s.SetTag(key, strconv.FormatBool(val))
	} else {
		s.SetExtra(key, val)
	}
}

// Labels sets labels for a logger in a batch
func (l *SentryLogger) SetAttrs(attrMap map[string]interface{}) {
	attrs.ApplyAttrs(l, attrMap)
}

// Req prepares a log to include http request information.
func (l *SentryLogger) Req(req *http.Request) Logger {
	if req == nil {
		return l
	}
	l.hub.Scope().SetRequest(req)
	return l
}

// Send the message that has been constructed.
func (m *SentryLoggerMsg) Send() {
	if m.entry.Level != sentry.LevelError {
		m.hub.CaptureEvent(m.entry)
		return
	}
	m.hub.WithScope(func(s *sentry.Scope) {
		s.SetExtras(m.entry.Extra)
		s.SetTags(m.entry.Tags)
		s.AddEventProcessor(func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			event.Message = m.entry.Message
			// keep only the first exception, because it already contains all levels
			// from the other error wrappers, we just need to extract the message
			// from the top level so we keep the detailed info.
			if len(event.Exception) > 0 {
				e := event.Exception[0]
				switch e.Type {
				// hide unnecessary types used in errors.Wrap calls
				case "*errors.withMessage", "*errors.withStack", "*errors.fundamental":
					e.Type = ""
				}
				e.Value = event.Exception[len(event.Exception)-1].Value
				event.Exception = []sentry.Exception{e}
			}
			return event
		})
		m.hub.CaptureException(m.err)
	})
}

// Str adds a tag to the message of type string.
func (m *SentryLoggerMsg) Str(key, val string) {
	if _, ok := m.allowedTags[key]; ok {
		m.entry.Tags[key] = val
		return
	}
	m.entry.Extra[key] = val
}

// I64 adds a tag to the message of type int64.
func (m *SentryLoggerMsg) I64(key string, val int64) {
	if _, ok := m.allowedTags[key]; ok {
		m.entry.Tags[key] = strconv.FormatInt(val, 10)
		return
	}
	m.entry.Extra[key] = val
}

// F64 adds a tag to the message of type float64.
func (m *SentryLoggerMsg) F64(key string, val float64) {
	if _, ok := m.allowedTags[key]; ok {
		m.entry.Tags[key] = strconv.FormatFloat(val, 'G', 9, 64)
		return
	}
	m.entry.Extra[key] = val
}

// Bool adds a tag to the message of type bool.
func (m *SentryLoggerMsg) Bool(key string, val bool) {
	if _, ok := m.allowedTags[key]; ok {
		m.entry.Tags[key] = strconv.FormatBool(val)
		return
	}
	m.entry.Extra[key] = val
}

func (m *SentryLoggerMsg) SetAttrs(attrMap map[string]interface{}) {
	attrs.ApplyAttrs(m, attrMap)
}
