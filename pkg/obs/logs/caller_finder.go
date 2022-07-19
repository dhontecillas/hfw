package logs

import (
	"runtime"
	"strings"
	"sync"
)

// this is extracted from the logrus package to be reused in
// different implementations: since we are wrapping other loggers
// we must implement the file location calling in this package
var (
	// qualified package name, cached at first use
	logsPackage string

	// Positions in the call stack when tracing to report the calling method
	minimumCallerDepth int

	// Used for caller information initialisation
	callerInitOnce sync.Once
)

const (
	maximumCallerDepth int = 25
	knownLogsFrames    int = 3
)

// getPackageName reduces a fully qualified function name to the package name
// There really ought to be to be a better way...
func getPackageName(f string) string {
	for {
		lastPeriod := strings.LastIndex(f, ".")
		lastSlash := strings.LastIndex(f, "/")
		if lastPeriod > lastSlash {
			f = f[:lastPeriod]
		} else {
			break
		}
	}

	return f
}

// GetCaller retrieves the name of the first non-logs calling function
func GetCaller() *runtime.Frame {
	// cache this package's fully-qualified name
	callerInitOnce.Do(func() {
		pcs := make([]uintptr, maximumCallerDepth)
		_ = runtime.Callers(0, pcs)

		// dynamic get the package name and the minimum caller depth
		for i := 0; i < maximumCallerDepth; i++ {
			funcName := runtime.FuncForPC(pcs[i]).Name()
			if strings.Contains(funcName, "GetCaller") {
				logsPackage = getPackageName(funcName)
				break
			}
		}

		minimumCallerDepth = knownLogsFrames
	})

	// Restrict the lookback frames to avoid runaway lookups
	pcs := make([]uintptr, maximumCallerDepth)
	depth := runtime.Callers(minimumCallerDepth, pcs)
	frames := runtime.CallersFrames(pcs[:depth])

	for f, again := frames.Next(); again; f, again = frames.Next() {
		pkg := getPackageName(f.Function)

		// If the caller isn't part of this package, we're done
		if pkg != logsPackage {
			return &f //nolint:scopelint
		}
	}

	// if we got here, we failed to find the caller's context
	return nil
}
