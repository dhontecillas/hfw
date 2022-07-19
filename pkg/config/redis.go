package config

import (
	"strconv"

	"github.com/spf13/viper"

	"github.com/dhontecillas/hfw/pkg/db"
)

const (
	confKeyRedisMasterHost string = "db.redis.master.host"
	confKeyRedisMasterPort string = "db.redis.master.port"
)

// ReadRedisConfig creates a db.RedisConfig from the
// environment or configuration file.
func ReadRedisConfig(confPrefix string) db.RedisConfig {
	portStr := viper.GetString(confPrefix + confKeyRedisMasterPort)
	port, err := strconv.ParseInt(portStr, 10, 32)
	if err != nil || port < 1 || port > 0xffff {
		port = 6379
	}
	host := viper.GetString(confPrefix + confKeyRedisMasterHost)
	if len(host) == 0 {
		host = "localhost"
	}
	return db.RedisConfig{
		Host: host,
		Port: int(port),
	}
}
