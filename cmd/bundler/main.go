package main

import (
	"fmt"
	"os"

	"github.com/spf13/viper"

	"github.com/dhontecillas/hfw/pkg/bundler"
	"github.com/dhontecillas/hfw/pkg/config"
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

	if err := config.InitConfig(confPrefix); err != nil {
		panic(err)
	}

	insConf := config.ReadInsightsConfig(confPrefix)
	if insConf == nil {
		panic("insConf is null")
	}
	appMetricDefs := metricsdefaults.HTTPDefaultMetricDefinitions()
	insB, insF := config.CreateInsightsBuilder(insConf, appMetricDefs)
	ins := insB()
	defer insF()

	bundler.ExecuteBundlerOperations(viper.GetViper(), ins.L, confPrefix)
}
