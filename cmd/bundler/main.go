package main

import (
	"github.com/spf13/viper"

	"github.com/dhontecillas/hfw/pkg/bundler"
	"github.com/dhontecillas/hfw/pkg/config"
	"github.com/dhontecillas/hfw/pkg/obs/metrics"
)

const (
	KeyConfPrefix string = "sendrules."
)

func main() {
	if err := config.InitConfig(KeyConfPrefix); err != nil {
		panic(err)
	}

	insConf := config.ReadInsightsConfig(KeyConfPrefix)
	if insConf == nil {
		panic("insConf is null")
	}
	appMetricDefs := make(metrics.Defs, 0)
	insB, insF := config.CreateInsightsBuilder(insConf, appMetricDefs)
	ins := insB()
	defer insF()

	bundler.ExecuteBundlerOperations(viper.GetViper(), ins.L, KeyConfPrefix)
}
