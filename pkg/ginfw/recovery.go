package ginfw

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime"
	"strings"

	"github.com/dhontecillas/hfw/pkg/extdeps"
	"github.com/dhontecillas/hfw/pkg/obs"

	"github.com/gin-gonic/gin"
)

var (
	dunno     = []byte("???")
	centerDot = []byte("·")
	dot       = []byte(".")
	slash     = []byte("/")
)

// RecoveryWithObs returns a middleware for a given writer that recovers
// from any panics and writes a 500 if there was one.
func RecoveryWithObs(ins *obs.Insighter) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			err := recover()
			if err == nil {
				return
			}

			reqIns := ins
			es, ok := c.Keys[extServicesKey].(*extdeps.ExternalServices)
			if ok && es.Insighter() != nil {
				// use the current request insighter if it is available
				reqIns = es.Insighter()
			}

			callStack := stack(3)
			httpRequest, _ := httputil.DumpRequest(c.Request, false)
			headers := strings.Split(string(httpRequest), "\r\n")
			// remove authorization data from the headers, to not log it
			for idx, header := range headers {
				current := strings.Split(header, ":")
				if current[0] == "Authorization" {
					headers[idx] = current[0] + ": *"
				}
			}

			lErr, ok := err.(error)
			if !ok {
				lErr = fmt.Errorf("PANIC")
			}
			msg := reqIns.L.ErrMsg(lErr, fmt.Sprintf("PANIC %s", lErr.Error()))
			msg.Str("headers", strings.Join(headers, "\r\n"))
			msg.Str("callstack", string(callStack))
			msg.Send()

			// Check for a broken connection, as it is not really a
			// condition that warrants a panic stack trace.
			if ne, ok := err.(*net.OpError); ok {
				if se, ok := ne.Err.(*os.SyscallError); ok {
					lowerErr := strings.ToLower(se.Error())
					if strings.Contains(lowerErr, "broken pipe") ||
						strings.Contains(lowerErr, "connection reset by peer") {
						_ = c.Error(err.(error)) // nolint: errcheck
						c.Abort()
					}
				}
			}
			c.AbortWithStatus(http.StatusInternalServerError)
		}()
		c.Next()
	}
}

// stack returns a nicely formatted stack frame, skipping skip frames.
func stack(skip int) []byte {
	buf := new(bytes.Buffer) // the returned data
	// As we loop, we open files and read them. These variables record the currently
	// loaded file.
	var lines [][]byte
	var lastFile string
	for i := skip; ; i++ { // Skip the expected number of frames
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		// Print this much at least.  If we can't find the source, it won't show.
		fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)
		if file != lastFile {
			data, err := ioutil.ReadFile(file)
			if err != nil {
				continue
			}
			lines = bytes.Split(data, []byte{'\n'})
			lastFile = file
		}
		fmt.Fprintf(buf, "\t%s: %s\n", function(pc), source(lines, line))
	}
	return buf.Bytes()
}

// source returns a space-trimmed slice of the n'th line.
func source(lines [][]byte, n int) []byte {
	n-- // in stack trace, lines are 1-indexed but our array is 0-indexed
	if n < 0 || n >= len(lines) {
		return dunno
	}
	return bytes.TrimSpace(lines[n])
}

// function returns, if possible, the name of the function containing the PC.
func function(pc uintptr) []byte {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return dunno
	}
	name := []byte(fn.Name())
	// The name includes the path name to the package, which is unnecessary
	// since the file name is already included.  Plus, it has center dots.
	// That is, we see
	//	runtime/debug.*T·ptrmethod
	// and want
	//	*T.ptrmethod
	// Also the package path might contains dot (e.g. code.google.com/...),
	// so first eliminate the path prefix
	if lastSlash := bytes.LastIndex(name, slash); lastSlash >= 0 {
		name = name[lastSlash+1:]
	}
	if period := bytes.Index(name, dot); period >= 0 {
		name = name[period+1:]
	}
	name = bytes.Replace(name, centerDot, dot, -1)
	return name
}
