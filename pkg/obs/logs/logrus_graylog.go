package logs

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
	graylog "gopkg.in/gemnasium/logrus-graylog-hook.v2"
)

const (
	prefixIgnore = "g-"
)

// CustomGraylogHook adds a new layer over gemnasium Graylog hook to handle specific
// cases when trigerred.
type CustomGraylogHook struct {
	*graylog.GraylogHook
	fieldPrefix string
}

// Fire is called when a log event is fired.
func (g *CustomGraylogHook) Fire(entry *logrus.Entry) error {
	// don't modify entry as it could be used by other hooks.
	// we need to make some changes to avoid indexing issues on Elasticsearch
	// change http.Request to a simple type and prefix any field names if the
	// option is given to avoid issues with dynamic types between different
	// library users.
	newData := make(map[string]interface{})
	for k, v := range entry.Data {
		newKey := k
		if g.fieldPrefix != "" && k != logrus.ErrorKey && !strings.HasPrefix(k, prefixIgnore) {
			newKey = fmt.Sprintf("%s-%s", g.fieldPrefix, k)
		}
		switch d := v.(type) {
		case *http.Request:
			newData[newKey] = newHTTPRequest(d)
		default:
			newData[newKey] = v
		}
	}

	newEntry := &logrus.Entry{
		Logger:  entry.Logger,
		Data:    newData,
		Time:    entry.Time,
		Level:   entry.Level,
		Message: entry.Message,
	}

	return g.GraylogHook.Fire(newEntry)
}

// NewCustomGraylogHook creates a hook to be added to an instance of logger.
func NewCustomGraylogHook(addr, fieldPrefix string) *CustomGraylogHook {
	h := &CustomGraylogHook{
		GraylogHook: graylog.NewAsyncGraylogHook(addr, nil),
		fieldPrefix: fieldPrefix,
	}

	return h
}

type httpRequest struct {
	Method        string      `json:"method"`
	Path          string      `json:"path"`
	QueryString   url.Values  `json:"queryString"`
	Proto         string      `json:"proto"`
	Header        http.Header `json:"header"`
	Body          string      `json:"body"`
	ContentLength int64       `json:"contentLength"`
	Host          string      `json:"host"`
	RemoteAddr    string      `json:"remoteAddr"`
}

func newHTTPRequest(r *http.Request) httpRequest {
	h := httpRequest{
		Method:        r.Method,
		Path:          r.URL.Path,
		QueryString:   r.URL.Query(),
		Proto:         r.Proto,
		Header:        r.Header,
		ContentLength: r.ContentLength,
		Host:          r.Host,
		RemoteAddr:    r.RemoteAddr,
	}
	h.Body = "NOT LOGGED"

	return h
}
