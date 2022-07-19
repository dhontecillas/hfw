package db

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	// we need pq in order to access a postgres database
	"github.com/dhontecillas/hfw/pkg/obs"
	_ "github.com/lib/pq"
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
	Name     string
	Host     string
	Port     int
	User     string
	Password string
}

// sqlDB implments the SQLDB interface
type sqlDB struct {
	master *sqlx.DB

	ins        *obs.Insighter
	connString string
}

// Master returns the connection to the master db
func (s *sqlDB) Master() *sqlx.DB {
	if s.master == nil {
		s.ins.L.Warn("master sqlx.DB connection is nil")
		s.connect()
	}
	return s.master
}

// Close
func (s *sqlDB) Close() {
	if s.master != nil {
		if err := s.master.Close(); err != nil {
			s.ins.L.Err(err, "closing connection")
		}
		s.master = nil
	}
}

// connect tries to connect to the database
func (s *sqlDB) connect() {
	masterDB, err := sqlx.Connect("postgres", s.connString)
	if err != nil {
		s.ins.L.Err(err, "cannot connect to server")
		panic("cannot connect to db server")
	}
	s.master = masterDB
}

// NewSQLDB creates a connection to a database.
// In case it cannot connect, it will panic.
func NewSQLDB(ins *obs.Insighter, conf *Config) SQLDB {
	if conf == nil {
		ins.L.Warn("No config for SQL DB provided")
		return nil
	}

	connString := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=disable",
		conf.Host, conf.Port, conf.Name, conf.User, conf.Password)
	sDB := &sqlDB{
		connString: connString,
		ins:        ins,
	}
	sDB.connect()
	return sDB
}
