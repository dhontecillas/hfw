/*
This is an example of how to use the observability library.

It expects to have these env vars set:

- SENTRY_DSN: a valid Sentry DSN to send logs to
- OBSEXAMPLE_LOGFILE
*/
package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/dhontecillas/hfw/pkg/obs"
	"github.com/dhontecillas/hfw/pkg/obs/logs"
	"github.com/dhontecillas/hfw/pkg/obs/metrics"
	"github.com/dhontecillas/hfw/pkg/obs/traces"
)

// Environment var the the example
const (
	EnvSentryDsn string = "SENTRY_DSN"
	EnvLogfile   string = "OBSEXAMPLE_LOGFILE"
)

func main() {
	fmt.Println("----------=== obs example ===----------")
	fmt.Println("")

	logBuilders := make([]logs.LoggerBuilderFn, 0, 4)
	logrusBuilder, logrusFlush := buildLogrusLogger()
	if logrusBuilder != nil {
		logBuilders = append(logBuilders, logrusBuilder)
		defer logrusFlush()
	}

	dsn := os.Getenv(EnvSentryDsn)
	if len(dsn) > 0 {
		sentryBuilder, sentryFlush, err := logs.NewSentryBuilder(&logs.SentryConf{
			Dsn:              dsn,
			AttachStacktrace: true,
			Environment:      "DEBUG",
			FlushTimeoutSecs: 4,
		})
		if err != nil {
			fmt.Printf("\n\nErr: %s\n", err.Error())
			panic(err.Error())
		}
		if sentryBuilder != nil {
			logBuilders = append(logBuilders, sentryBuilder)
			defer sentryFlush()
		} else {
			fmt.Printf("Sentry builder is NIL\n")
		}
	}

	// we can have several log builders and wrap them to have logs sent to
	// different places
	logBuilder := logs.NewMultiLoggerBuilder(logBuilders...)

	startupLogger := logBuilder()
	// example of sending metrics to multiple meters
	mPromBuilder := buildPrometheusMeterBuilder(startupLogger)
	mNopBuilder, err := metrics.NewNopMeterBuilder()
	if err != nil {
		return
	}
	meterBuilder, err := metrics.NewMultiMeterBuilder(startupLogger, mPromBuilder, mNopBuilder)
	if err != nil {
		return
	}

	tracerBuilder := traces.NewNopTracerBuilder()

	// get the builder function for the Insights instance
	insBuilder := obs.NewInsighterBuilder(
		[]obs.TagDefinition{
			obs.TagDefinition{
				Name:    "path",
				TagType: obs.TagTypeStr,
				ToL:     true,
				ToM:     true,
				ToT:     false,
			},
			obs.TagDefinition{
				Name:    "req_id",
				TagType: obs.TagTypeI64,
				ToL:     true,
				ToM:     true,
				ToT:     false,
			},
		},
		logBuilder, meterBuilder, tracerBuilder)

	maxConcurrent := 200
	maxLatency := 500
	minLatency := 70

	endChan := make(chan interface{}, maxConcurrent)
	concurrent := 0

	fakePaths := []string{"marge", "homer", "bart", "lisa", "maggie"}
	fakeMethods := []string{"GET", "POST", "GET", "PUT", "GET", "GET", "POST", "DELETE"}
	fakeVersions := []int{1, 2, 3}

	cntr := 0
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for ; ; cntr++ {
		if concurrent >= maxConcurrent {
			// wait for half of the fake goroutines to finish
			for ii := 0; ii < maxConcurrent/2; ii++ {
				<-endChan
				concurrent--
			}
		}

		concurrent++
		ins := insBuilder()

		// create a fake path
		path := fmt.Sprintf("/v%d/%s", cntr%len(fakeVersions),
			fakePaths[cntr%len(fakePaths)])

		var fakeLatency int64 = int64(minLatency) + int64(r.Uint64()%uint64(maxLatency-minLatency))
		go newFakeRequest(endChan, ins, int64(cntr), path, fakeMethods[cntr%len(fakeMethods)],
			fakeLatency, 2)
		time.Sleep(time.Millisecond)
		// stop it
		// break
	}
}

func buildLogrusLogger() (logs.LoggerBuilderFn, func()) {
	outFileName := os.Getenv(EnvLogfile)
	if len(outFileName) == 0 {
		outFileName = "./examples/obs_example/compose/tmp/example_log.txt"
	}
	logrusConf := logs.LogrusConf{
		OutFileName: outFileName,
	}
	lB, lBFlush, err := logs.NewLogrusBuilder(&logrusConf)
	if err != nil {
		fmt.Printf("cannot build logrus : %s\n\n", err)
		return nil, nil
	}
	return lB, lBFlush
}

func buildPrometheusMeterBuilder(l logs.Logger) metrics.MeterBuilderFn {
	pConf := metrics.PrometheusConfig{
		ServerPort: ":9876",
		ServerPath: "/prom_metrics",
		MetricDefinitions: []metrics.Def{
			metrics.Def{
				Name:       "requests",
				MetricType: metrics.MetricTypeMonotonicCounter,
				Labels:     []string{"path", "verb"},
			},
			metrics.Def{
				Name:       "concurrent_requests",
				MetricType: metrics.MetricTypeUpDownCounter,
				Labels:     []string{"verb"},
			},
		},
	}
	pmBuilder, err := metrics.NewPrometheusMeterBuilder(l, &pConf)
	if err != nil {
		l.Err(err, "Cannot create meter")
		return nil
	}

	// we need to server the metrics in a separate port:
	metrics.Serve(&pConf)

	return pmBuilder
}

func newFakeRequest(endChan chan interface{}, ins *obs.Insighter,
	reqID int64, path string, method string,
	millis int64, spawnSubprocess int64) {

	ins.Str("path", path)
	ins.I64("req_id", reqID)
	ins.M.Str("path", path)
	ins.M.Str("verb", method)
	ins.M.Inc("concurrent_requests")
	ins.M.Inc("requests")

	tr := ins.T.Start("newFakeRequest")
	defer tr.End()
	tr.Str("lol", "bah")
	ins.L.Str("trace_id", tr.TraceID())

	st := time.Now()
	et := st.Add(time.Duration(time.Millisecond) * time.Duration(millis))

	ins.L.Info(fmt.Sprintf("Request %d start %v , end %v", reqID, st, et))
	for ; et.After(time.Now()); time.Sleep(time.Millisecond * time.Duration(50)) {
		if spawnSubprocess > 0 {
			spawnSubprocess--
			subIns := ins.CloneWith(ins.L.Clone(), ins.M, tr)
			go newBgProc(subIns)
		}
		ins.M.Inc("requests")
	}
	// signal the end of the goroutine
	ins.M.Dec("concurrent_requests")
	endChan <- nil
}

func newBgProc(ins *obs.Insighter) {
	tr := ins.T.Start("newBgProc")
	defer tr.End()
	time.Sleep(time.Millisecond * time.Duration(20))
}
