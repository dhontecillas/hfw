package main

import (
	"fmt"
	"os"

	"github.com/dhontecillas/hfw/pkg/bundler"
	"github.com/dhontecillas/hfw/pkg/config"
	"github.com/dhontecillas/hfw/pkg/db"
	metricsdefaults "github.com/dhontecillas/hfw/pkg/obs/metrics/defaults"
)

const (
	EnvKeyBundlerConfPrefix string = "HFW_BUNDLER_PREFIX"
)

func main() {
	confPrefix := os.Getenv(EnvKeyBundlerConfPrefix)
	if len(confPrefix) == 0 {
		fmt.Printf("the HFW_BUNDLER_PREFIX must be set and should be something like 'yourappprefix.'\n")
		return
	}

	cldr, err := config.InitConfig(confPrefix)
	if err != nil {
		panic(err)
	}

	insConf := config.ReadInsightsConfig(cldr)
	if insConf == nil {
		panic("insConf is null")
	}
	appMetricDefs := metricsdefaults.HTTPDefaultMetricDefinitions()
	insB, insF := config.CreateInsightsBuilder(insConf, appMetricDefs)
	ins := insB()
	defer insF()

	bundlerConfLoader, err := cldr.Section([]string{"bundler"})
	if err != nil {
		panic("cannot find bundler configuration")
	}
	dbConfLoader, err := cldr.Section([]string{"db", "sql", "master"})
	if err != nil {
		panic("cannot find db configuration")
	}
	var bundlerConf config.BundlerConfig
	if err := bundlerConfLoader.Parse(&bundlerConf); err != nil {
		panic("cannot load bundler config")
	}
	var dbConf db.Config
	if err := dbConfLoader.Parse(&dbConf); err != nil {
		panic("cannot load db config")
	}
	bundler.ExecuteBundlerOperations(&bundlerConf, &dbConf, ins.L)
}
