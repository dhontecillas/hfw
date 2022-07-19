package db

import (
	"fmt"
)

// RedisConfig contains the configuration to access a
// Redis instance.
type RedisConfig struct {
	Host string
	Port int
}

// Address returns the address for a redis instance.
func (rc *RedisConfig) Address() string {
	return fmt.Sprintf("%s:%d", rc.Host, rc.Port)
}
