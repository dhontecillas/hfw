package config

import (
	"github.com/dhontecillas/hfw/pkg/db"
	"github.com/dhontecillas/hfw/pkg/obs"
)

/*
const (
	confKeyDBMasterName string = "db.sql.master.name"
	confKeyDBMasterHost string = "db.sql.master.host"
	confKeyDBMasterPort string = "db.sql.master.port"
	confKeyDBMasterUser string = "db.sql.master.user"
	confKeyDBMasterPass string = "db.sql.master.pass"
)
*/

type SQLConfig struct {
	Master      db.Config `json:"master"`
	ReadReplica db.Config `json:"readreplica"`
}

type DBConfig struct {
	SQL SQLConfig `json:"sql"`
}

// ReadSQLDBConfig reads the configuration for the database
// using the application configuration prefix.
func ReadSQLDBConfig(cldr ConfLoader) (*db.Config, error) {
	var err error
	cldr, err = cldr.Section([]string{"db"})
	if err != nil {
		return nil, err
	}
	var conf DBConfig
	if err := cldr.Parse(conf); err != nil {
		return nil, err
	}
	if err = conf.SQL.Master.Validate(); err != nil {
		return nil, err
	}
	return &conf.SQL.Master, err
}

// CreateSQLDB creates a new database connection.
func CreateSQLDB(ins *obs.Insighter, conf *db.Config) db.SQLDB {
	return db.NewSQLDB(ins, conf)
}
