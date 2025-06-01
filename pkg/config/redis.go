package config

import (
	"github.com/dhontecillas/hfw/pkg/db"
)

const (
	confKeyRedisMasterHost string = "db.redis.master.host"
	confKeyRedisMasterPort string = "db.redis.master.port"
)

type RedisConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

func (c *RedisConfig) Validate() error {
	if c.Host == "" {
		c.Host = "localhost"
	}
	if c.Port < 0 || c.Port > 65536 {
		return ErrBadPortNumber
	}
	if c.Port == 0 {
		c.Port = 6379
	}

	return nil
}

// ReadRedisConfig creates a db.RedisConfig from the
// environment or configuration file.
func ReadRedisConfig(cldr ConfLoader) *db.RedisConfig {
	// TODO: allow to return numbers
	var err error
	var conf db.RedisConfig
	cldr, err = cldr.Section([]string{"db", "redis", "master"})
	if err != nil {
		return nil
	}
	err = cldr.Parse(&conf)
	if err != nil {
		return nil
	}
	return &conf
}
