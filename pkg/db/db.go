package db

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	// we need pq in order to access a postgres database
	"github.com/dhontecillas/hfw/pkg/obs"
	_ "github.com/lib/pq"
)

var (
	ErrMissingDBName = fmt.Errorf("missing db name")
	ErrMissingDBHost = fmt.Errorf("missing db host")
	ErrMissingDBUser = fmt.Errorf("missing db user")
)

// SQLDB contains the connection to a master
// database and a way to shutdown the connection
// using Close
type SQLDB interface {
	Master() *sqlx.DB
	Close()
}

// Config contains the basic DB configuration params.
type Config struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
}

func (c *Config) Validate() error {
	if c.Name == "" {
		return ErrMissingDBName
	}
	if c.Host == "" {
		return ErrMissingDBHost
	}
	if c.Port == 0 {
		c.Port = 5432
	}
	if c.User == "" {
		return ErrMissingDBUser
	}
	return nil
}

// sqlDB implments the SQLDB interface
type sqlDB struct {
	master *sqlx.DB

	ins        *obs.Insighter
	connString string
}

// Master returns the connection to the master db. It might
// be nil in case it cannot connect to the database
// TODO: we should return an error if this can fail
func (s *sqlDB) Master() *sqlx.DB {
	if s.master == nil {
		s.ins.L.Warn("master sqlx.DB connection is nil", nil)
		if err := s.connect(); err != nil {
			return nil
		}
	}
	return s.master
}

// Close
func (s *sqlDB) Close() {
	if s.master != nil {
		if err := s.master.Close(); err != nil {
			s.ins.L.Err(err, "closing connection", nil)
		}
		s.master = nil
	}
}

// connect tries to connect to the database
func (s *sqlDB) connect() error {
	masterDB, err := sqlx.Connect("postgres", s.connString)
	if err != nil {
		s.ins.L.Err(err, "cannot connect to server", nil)
		return fmt.Errorf("cannot connect to db server")
	}
	s.master = masterDB
	return nil
}

// NewSQLDB creates a connection to a database.
func NewSQLDB(ins *obs.Insighter, conf *Config) SQLDB {
	if conf == nil {
		ins.L.Warn("No config for SQL DB provided", nil)
		return nil
	}

	connString := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=disable",
		conf.Host, conf.Port, conf.Name, conf.User, conf.Password)
	sDB := &sqlDB{
		connString: connString,
		ins:        ins,
	}
	if err := sDB.connect(); err != nil {
		// we might be able to connect later
		ins.L.Err(err, "cannot connect to db server", map[string]interface{}{
			"conn_string": connString,
		})
	}
	return sDB
}
