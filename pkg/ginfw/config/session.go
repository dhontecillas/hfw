package ginfwconfig

import (
	"fmt"

	"github.com/spf13/viper"

	"github.com/dhontecillas/hfw/pkg/db"
	"github.com/dhontecillas/hfw/pkg/ginfw/web/session"
	"github.com/dhontecillas/hfw/pkg/obs"
)

const (
	confKeySessionRedisMaxIdle  string = "ginfw.session.redis.maxidle"
	confKeySessionRedisHost     string = "ginfw.session.redis.host"
	confKeySessionRedisPassword string = "ginfw.session.redis.password"
	confKeySessionSecretKeyPair string = "ginfw.session.secretkeypair"

	confDefaultSessionRedisMaxIdle  string = "10"
	confDefaultSessionRedisHost     string = "localhost:6379"
	confDefaultSessionRedisPassword string = ""

	confKeySessionCSRFSecret string = "ginfw.session.csrfsecret"
	confKeySessionIsDevelop  string = "ginfw.session.develop"

	confDefaultSessionIsDevelop string = "false"
)

// ReadSessionConf reads the required configuration to have
// a Seesion insttance.
func ReadSessionConf(ins *obs.Insighter, confPrefix string,
	redisConf *db.RedisConfig) (*session.Conf, error) {

	secretKeyPair := viper.GetString(confPrefix + confKeySessionSecretKeyPair)
	if len(secretKeyPair) == 0 {
		msg := fmt.Sprintf("cannot read required config value: %s",
			confKeySessionSecretKeyPair)
		err := fmt.Errorf(msg)
		ins.L.ErrMsg(err, msg)
		return nil, err
	}
	CSRFSecret := viper.GetString(confPrefix + confKeySessionCSRFSecret)
	if len(CSRFSecret) == 0 {
		msg := fmt.Sprintf("cannot read required config value: %s",
			confKeySessionCSRFSecret)
		err := fmt.Errorf(msg)
		ins.L.ErrMsg(err, msg)
		return nil, err
	}

	viper.SetDefault(confKeySessionRedisMaxIdle, confDefaultSessionRedisMaxIdle)
	if redisConf != nil {
		viper.SetDefault(confKeySessionRedisHost, redisConf.Address())
	} else {
		viper.SetDefault(confKeySessionRedisHost, confDefaultSessionRedisHost)
	}
	viper.SetDefault(confKeySessionRedisPassword, confDefaultSessionRedisPassword)
	viper.SetDefault(confKeySessionIsDevelop, confDefaultSessionIsDevelop)

	return &session.Conf{
		RedisConf: session.RedisConf{
			MaxIdleConnections: viper.GetInt(confKeySessionRedisMaxIdle),
			Host:               viper.GetString(confKeySessionRedisHost),
			Password:           viper.GetString(confKeySessionRedisPassword),
			SecretKeyPair:      secretKeyPair,
		},
		CsrfSecret: CSRFSecret,
		IsDevelop:  false,
	}, nil
}
