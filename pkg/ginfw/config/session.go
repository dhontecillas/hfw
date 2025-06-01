package ginfwconfig

import (
	"fmt"

	"github.com/dhontecillas/hfw/pkg/config"
	"github.com/dhontecillas/hfw/pkg/db"
	"github.com/dhontecillas/hfw/pkg/ginfw/web/session"
	"github.com/dhontecillas/hfw/pkg/obs"
)

const (
	confDefaultSessionRedisMaxIdle        = 10
	confDefaultSessionRedisHost    string = "localhost:6379"
)

type GinSessionRedisConfig struct {
	Host     string `json:"host"`
	Password string `json:"password"`
	MaxIdle  int    `json:"maxidle"`
}

func (c *GinSessionRedisConfig) Validate() error {
	if c.MaxIdle <= 0 {
		c.MaxIdle = confDefaultSessionRedisMaxIdle
	}
	return nil
}

type GinSessionConfig struct {
	Redis            GinSessionRedisConfig `json:"redis"`
	CSRFSecret       string                `json:"csrfsecret"`
	SecretKeyPair    string                `json:"secretkeypair"`
	SessionIsDevelop bool                  `json:"develop"`
}

func (c *GinSessionConfig) Validate() error {
	err := c.Redis.Validate()
	if err != nil {
		return err
	}
	if c.SecretKeyPair == "" {
		return fmt.Errorf("missing session 'secretkeypair'")
	}
	return nil
}

// ReadSessionConf reads the required configuration to have
// a Seesion insttance.
func ReadSessionConf(ins *obs.Insighter, cldr config.ConfLoader,
	redisConf *db.RedisConfig) (*session.Conf, error) {

	cldr, err := cldr.Section([]string{"ginfw", "session"})
	if err != nil {
		ins.L.Err(err, "cannot read ginfw session", nil)
		return nil, err
	}
	var conf GinSessionConfig
	if err := cldr.Parse(&conf); err != nil {
		ins.L.Err(err, "cannot parse ginfw session", nil)
		return nil, err
	}
	if err := conf.Validate(); err != nil {
		ins.L.Err(err, "cannot validate ginfw session", nil)
		return nil, err
	}

	if conf.Redis.Host == "" {
		if redisConf != nil {
			conf.Redis.Host = redisConf.Address()
		} else {
			conf.Redis.Host = confDefaultSessionRedisHost
		}
	}

	return &session.Conf{
		RedisConf: session.RedisConf{
			MaxIdleConnections: conf.Redis.MaxIdle,
			Host:               conf.Redis.Host,
			Password:           conf.Redis.Password,
			SecretKeyPair:      conf.SecretKeyPair,
		},
		CsrfSecret: conf.CSRFSecret,
		IsDevelop:  conf.SessionIsDevelop,
	}, nil
}
