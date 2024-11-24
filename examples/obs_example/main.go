/*
This is an example of how to use the observability library.

It expects to have these env vars set:

- SENTRY_DSN: a valid Sentry DSN to send logs to
- OBSEXAMPLE_LOGFILE
*/
package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/dhontecillas/hfw/pkg/obs"
	"github.com/dhontecillas/hfw/pkg/obs/attrs"
	"github.com/dhontecillas/hfw/pkg/obs/httpobs"
	"github.com/dhontecillas/hfw/pkg/obs/logs"
	"github.com/dhontecillas/hfw/pkg/obs/metrics"
	metattrs "github.com/dhontecillas/hfw/pkg/obs/metrics/attrs"
	metricsdefaults "github.com/dhontecillas/hfw/pkg/obs/metrics/defaults"
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

	metricDefs := metricsdefaults.HTTPDefaultMetricDefinitions()
	metricDefs, _ = metricDefs.CleanUp()

	// we can have several log builders and wrap them to have logs sent to
	// different places
	logBuilder := logs.NewMultiLoggerBuilder(logBuilders...)

	startupLogger := logBuilder()
	// example of sending metrics to multiple meters
	mPromBuilder := buildPrometheusMeterBuilder(startupLogger, metricDefs)

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
	insBuilder := obs.NewInsighterBuilder(metricDefs, logBuilder,
		meterBuilder, tracerBuilder)

	ins := insBuilder()

	// create a test server with delays between 70 and 500 ms
	s := httptest.NewServer(newFakeHandler(ins, 70, 500))
	defer s.Close()

	ctx, cancel := context.WithCancel(context.Background())
	// launch 20 clients, making requests with a delay of 1 second
	// per request and a jitter of about 1 second
	launchClients(ctx, ins, s.URL, 20, time.Second, time.Second)

	fmt.Printf("waiting for 10 secs ...")
	time.Sleep(time.Second * 10)
	cancel()
	fmt.Printf("shutting down ... ")
	time.Sleep(time.Second * 2)
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

func buildPrometheusMeterBuilder(l logs.Logger,
	mdefs metrics.MetricDefinitionList) metrics.MeterBuilderFn {

	pConf := metrics.PrometheusConfig{
		ServerPort: ":9876",
		ServerPath: "/prom_metrics",
	}
	pmBuilder, err := metrics.NewPrometheusMeterBuilder(l, &pConf, mdefs)
	if err != nil {
		l.Err(err, "Cannot create meter", map[string]interface{}{
			"port": pConf.ServerPort,
			"path": pConf.ServerPath,
		})
		return nil
	}

	// we need to server the metrics in a separate port:
	metrics.Serve(&pConf)

	return pmBuilder
}

// the matched route cannot be extracted from the request, but from
// the router that is handling it
func SetReqAttrs(r *http.Request, route string, at attrs.Attributable) {
	if f, e := httpobs.ExtractTelemetryFields(r); e != nil {
		f[metattrs.AttrHTTPRoute] = route
		at.SetAttrs(f)
	}
}

func newFakeHandler(ins *obs.Insighter, minLatency time.Duration, maxLatency time.Duration) http.HandlerFunc {

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	latencyExtra := maxLatency - minLatency
	return func(w http.ResponseWriter, r *http.Request) {
		lat := minLatency + time.Duration(float64(latencyExtra)*rnd.Float64())
		// TODO: add background processes to see traces
		// bgProcs := rnd.Intn(3)

		time.Sleep(lat)
		h := w.Header()
		h.Add("X-Fake", "fake value")
		w.WriteHeader(200)
		w.Write([]byte("{'foo': bar}"))
	}
}

func launchClients(ctx context.Context, ins *obs.Insighter, URL string,
	numClients int, period time.Duration, jitter time.Duration) {

	for i := 0; i < numClients; i++ {
		go fakeClient(ctx, ins, period, jitter, URL)
	}
}

// fakeClient keeps launching requests to a server randomly
// until the context is cancelled
//
// period -> the time it waits after sending a request
func fakeClient(ctx context.Context, ins *obs.Insighter, period time.Duration,
	jitter time.Duration, host string) {

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	tk := time.NewTicker(period)
	for {
		select {
		case <-tk.C:
			sendClientRequest(ins, host, rnd)
			time.Sleep(period + time.Duration(float64(jitter)*rnd.Float64()))
		case <-ctx.Done():
			break
		}
	}
}

func sendClientRequest(ins *obs.Insighter, host string, rnd *rand.Rand) {
	methods := []string{
		"GET",
		"POST",
		"PUT",
		"GET",
		"DELETE",
		"POST",
		"GET",
	}
	paths := []string{
		"/marge",
		"/homer",
		"/bart",
	}
	bodies := []string{
		"{ 'foo': 'bar' }",
		"{ 'a': 23, 'b': 12}",
	}

	m := methods[rnd.Intn(len(methods))]
	var b io.Reader
	b = http.NoBody
	if m == "POST" || m == "PUT" {
		b = bytes.NewReader([]byte(bodies[rnd.Intn(len(bodies))]))
	}
	path := paths[rnd.Intn(len(paths))]
	url := host + path
	r, err := http.NewRequest(m, url, b)
	if err != nil {
		ins.L.Info("cannot create request", map[string]interface{}{
			"method": m,
			"url":    url,
			"err":    err,
		})
	}

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		ins.L.Info("cannot execute request", map[string]interface{}{
			"method": m,
			"url":    url,
			"err":    err,
		})
	} else {
		ins.L.Info("got response", map[string]interface{}{
			"resp": fmt.Sprintf("%+v", resp),
		})
	}
}

func newBgProc(ins *obs.Insighter) {
	tr := ins.T.Start("newBgProc")
	defer tr.End()
	time.Sleep(time.Millisecond * time.Duration(20))
}
