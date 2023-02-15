package testing

import (
	"os"
	"strconv"

	"github.com/dhontecillas/hfw/pkg/db"
	"github.com/dhontecillas/hfw/pkg/extdeps"
	"github.com/dhontecillas/hfw/pkg/mailer"
	"github.com/dhontecillas/hfw/pkg/notifications"
	"github.com/dhontecillas/hfw/pkg/obs"
	"github.com/dhontecillas/hfw/pkg/obs/logs"
	"github.com/dhontecillas/hfw/pkg/obs/metrics"
	"github.com/dhontecillas/hfw/pkg/obs/traces"
)

// Environment variables for building the external
// dependencies for the testing environment.
const (
	TestDBPort     = "TESTDB_PORT"
	TestDBName     = "TESTDB_NAME"
	TestDBHost     = "TESTDB_HOST"
	TestDBUser     = "TESTDB_USER"
	TestDBPassword = "TESTDB_PASSWORD"
)

func testDBConfig() *db.Config {
	sqlConf := db.Config{
		Name:     "hfwtest",
		Host:     "127.0.0.1",
		Port:     5430,
		User:     "hfwtest",
		Password: "test",
	}

	port, err := strconv.Atoi(os.Getenv(TestDBPort))
	if err == nil {
		sqlConf.Port = port
	}
	name := os.Getenv(TestDBName)
	if len(name) > 0 {
		sqlConf.Name = name
	}
	host := os.Getenv(TestDBHost)
	if len(host) > 0 {
		sqlConf.Host = host
	}
	user := os.Getenv(TestDBUser)
	if len(user) > 0 {
		sqlConf.User = user
	}
	password := os.Getenv(TestDBPassword)
	if len(password) > 0 {
		sqlConf.Password = password
	}

	return &sqlConf
}

// BuildExternalServices createes the external services for
// being used in tests.
func BuildExternalServices() *extdeps.ExternalServices {
	logFn := logs.NewNopLoggerBuilder()
	meterFn, _ := metrics.NewNopMeterBuilder()
	tracerFn := traces.NewNopTracerBuilder()
	insBuilderFn := obs.NewInsighterBuilder([]obs.TagDefinition{},
		logFn, meterFn, tracerFn)

	sqlConf := testDBConfig()

	mailer := mailer.NewNopMailer()
	composer := notifications.NewFileSystemComposer("./pkg/notifications")
	flushFn := func() {}
	return extdeps.NewExternalServices(insBuilderFn, flushFn, mailer,
		db.NewSQLDB(insBuilderFn(), sqlConf), composer)
}
