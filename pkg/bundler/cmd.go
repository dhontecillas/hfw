// Package bundler takes care of collecting file assets stored
// in different packages and putting them in a single place that
// might be accessible to the app or a third party static file
// server.
//
// These are the kind of assets that can take care of collecting:
//   - migration files
//   - html templates
//   - notification templates
package bundler

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	// we need to import postgres to initialize the db for migrations
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	// we need to import file to be able to use migrations from file
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/dhontecillas/hfw/pkg/config"
	"github.com/dhontecillas/hfw/pkg/db"
	"github.com/dhontecillas/hfw/pkg/obs/logs"
)

var (
	ErrBundlerNoSection = errors.New("no 'bundler' section")
	ErrBundlerParse     = errors.New("cannot parse 'bundler' section")
	ErrBundlerValidate  = errors.New("cannot validate 'bundler' section")
)

type BundlerMigrationsConfig struct {
	Collect bool     `json:"collect"`
	Dst     string   `json:"dst"`
	Scan    []string `json:"scan"`    // directories to scan for migrations
	Migrate string   `json:"migrate"` // migrate version
}

func (c *BundlerMigrationsConfig) Validate() error {
	// if Dst == "" there is no bundling enabled
	if c.Dst != "" {
		dst, err := filepath.Abs(c.Dst)
		if err != nil {
			return err
		}
		c.Dst = dst
	}
	// TODO: validate that the scan directoriese exist
	return nil
}

type BundlerPackConfig struct {
	// where to place the bundle files
	Dst string `json:"dst"`
	// from where the files are picked to be packed with the bundler:
	Srcs []string `json:"srcs"`
	// variant of the config file to be put inside the bundle:
	Variant string `json:"variant"`
}

func (c *BundlerPackConfig) Validate() error {
	if c.Dst != "" {
		dst, err := filepath.Abs(c.Dst)
		if err != nil {
			return err
		}
		c.Dst = dst
	}

	return nil
}

type BundlerConfig struct {
	Pack       BundlerPackConfig       `json:"pack"`
	Migrations BundlerMigrationsConfig `json:"migrations"`
}

func (c *BundlerConfig) Validate() error {
	if err := c.Pack.Validate(); err != nil {
		return err
	}
	if err := c.Migrations.Validate(); err != nil {
		return err
	}
	return nil
}

func NewBundlerConfig(cldr config.ConfLoader) (*BundlerConfig, error) {
	cldr, err := cldr.Section([]string{"bundler"})
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrBundlerNoSection,
			err.Error())
	}
	var cnf BundlerConfig
	if err := cldr.Parse(&cnf); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrBundlerParse, err.Error())
	}
	if err := cnf.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrBundlerValidate, err.Error())
	}
	return &cnf, nil
}

// ExecuteBundlerOperations parses the command line and environment
// to find operations that the bundler should execute: collect migrations,
// collect static files, run migrations, etc..
// This function is a helper to be able to run those operations from
// the same server executable file.
func ExecuteBundlerOperations(conf *BundlerConfig, dbConf *db.Config, l logs.Logger) {

	if conf.Migrations.Collect {
		if err := UpdateMigrations(conf.Migrations.Dst, conf.Migrations.Scan, l); err != nil {
			l.Err(err, "cannot update migration from config", nil)
		}
	}

	if conf.Migrations.Migrate != "" {
		err := ApplyMigrationsFromConfig(&conf.Migrations, dbConf, l)
		if err != nil {
			l.Err(err, "cannot apply migration", nil)
		}
	}

	if conf.Pack.Dst != "" {
		projDir, err := os.Getwd()
		if err != nil {
			l.Err(err, "cannot read working directory", nil)
			return
		}
		//  get the bundle variant
		//
		// TODO: we might have a conf variant set! we must use
		// the hardcoded "prod" only as a fallback
		err = PrepareBundle(projDir, conf.Pack.Dst, conf.Pack.Srcs, conf.Pack.Variant)
		if err != nil {
			l.Err(err, "cannot prepare bundle", nil)
			return
		}
	} else {
		l.Info("not bundling enabled", nil)
	}
}

// ApplyMigrationsFromConfig reads the migrations firectory from config and
// applies them.
func ApplyMigrationsFromConfig(conf *BundlerMigrationsConfig, dbConf *db.Config, l logs.Logger) error {
	if conf.Migrate == "" {
		l.Info("no migration to apply", nil)
		return nil
	}
	if dbConf == nil {
		return fmt.Errorf("no sql db config")
	}
	fURL := fmt.Sprintf("file://%s", conf.Dst)
	pgURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		dbConf.User, dbConf.Password, dbConf.Host, dbConf.Port, dbConf.Name)
	mig, err := migrate.New(fURL, pgURL)
	if err != nil {
		return fmt.Errorf("cannot apply %s migration to %s: %s",
			fURL, pgURL, err.Error())
	}

	if conf.Migrate == "up" {
		if err := mig.Up(); err != nil && err != migrate.ErrNoChange {
			return err
		}
	} else if conf.Migrate == "down" {
		if err := mig.Down(); err != nil {
			return err
		}
	} else {
		// read the value as a number, we run ALWAYS on 64 bit,
		// because migrations numbers are that long
		verNum64, err := strconv.ParseUint(conf.Migrate, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot read migration version %s: %s",
				conf.Migrate, err.Error())
		}
		verNumber := uint(verNum64)
		if err := mig.Migrate(verNumber); err != nil {
			return err
		}
	}
	return nil
}
