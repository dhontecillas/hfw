package config

import (
	"testing"

	"github.com/spf13/viper"
)

func TestInitSendGrid(t *testing.T) {
	confPrefix := "test."
	confKey := confPrefix + confKeySendgridKey
	conf, err := configSendGrid(confPrefix)
	if err == nil {
		t.Errorf("Expected error for unset key: %s", confKey)
		return
	}
	if len(conf.Key) != 0 {
		t.Errorf("Expected conf.Key to be empty")
		return
	}
	viper.Set(confKey, "FOO")
	conf, err = configSendGrid(confPrefix)
	if err != nil {
		t.Errorf("err is not nil: %s", err.Error())
		return
	}
	if len(conf.Key) == 0 {
		t.Errorf("bad configuration read")
		return
	}
}
