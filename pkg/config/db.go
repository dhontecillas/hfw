package config

import (
	"fmt"
	"strconv"

	"github.com/spf13/viper"

	"github.com/dhontecillas/hfw/pkg/db"
	"github.com/dhontecillas/hfw/pkg/obs"
)

const (
	confKeyDBMasterName string = "db.sql.master.name"
	confKeyDBMasterHost string = "db.sql.master.host"
	confKeyDBMasterPort string = "db.sql.master.port"
	confKeyDBMasterUser string = "db.sql.master.user"
	confKeyDBMasterPass string = "db.sql.master.pass"
)

// ReadSQLDBConfig reads the configuration for the database
// using the application configuration prefix.
func ReadSQLDBConfig(confPrefix string) (*db.Config, error) {
	var err error
	conf := &db.Config{}

	portStr := viper.GetString(confPrefix + confKeyDBMasterPort)
	if len(portStr) == 0 {
		portStr = "5432"
	}
	port, err := strconv.ParseInt(portStr, 10, 32)
	if err != nil {
		return nil, err
	}
	conf.Port = int(port)

	conf.Name = viper.GetString(confPrefix + confKeyDBMasterName)
	conf.Host = viper.GetString(confPrefix + confKeyDBMasterHost)
	conf.User = viper.GetString(confPrefix + confKeyDBMasterUser)
	conf.Password = viper.GetString(confPrefix + confKeyDBMasterPass)

	if len(conf.Name) == 0 {
		return nil, fmt.Errorf("missing required DB config: %s",
			confPrefix+confKeyDBMasterName)
	}
	if len(conf.Host) == 0 {
		return nil, fmt.Errorf("missing required DB config: %s",
			confPrefix+confKeyDBMasterHost)
	}
	if len(conf.User) == 0 {
		return nil, fmt.Errorf("missing required DB config: %s", confPrefix+confKeyDBMasterUser)
	}
	if len(conf.Password) == 0 {
		return nil, fmt.Errorf("missing required DB config: %s", confPrefix+confKeyDBMasterPass)
	}
	return conf, err
}

// CreateSQLDB creates a new database connection.
func CreateSQLDB(ins *obs.Insighter, conf *db.Config) db.SQLDB {
	return db.NewSQLDB(ins, conf)
}
