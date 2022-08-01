// Package bundler takes care of collecting file assets stored
// in different packages and putting them in a single place that
// might be accessible to the app or a third party static file
// server.
//
// These are the kind of assets that can take care of collecting:
//  - migration files
//  -
//
package bundler

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	// we need to import postgres to initialize the db for migrations
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	// we need to import file to be able to use migrations from file
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/viper"

	"github.com/dhontecillas/hfw/pkg/config"
	"github.com/dhontecillas/hfw/pkg/obs/logs"
)

const (
	// KeyBundlerCollectMigrations is the config key for a boolean value that enables
	// collecting the database migration files.
	KeyBundlerCollectMigrations = "bundler.migrations.collect"

	// KeyBundlerCollectMigrationsDstDir is the config key for a string value that
	// tells the bundler where to place the collected migrations.
	KeyBundlerCollectMigrationsDstDir = "bundler.migrations.dst"

	// KeyBundlerCollectMigrationsScanDirs  is the config key for migrations source dir.
	KeyBundlerCollectMigrationsScanDirs = "bundler.migrations.scan"

	// KeyBundlerMigrate is the config key for knowing if migrations should be applied.
	KeyBundlerMigrate = "bundler.migrations.migrate"
	// KeyBundlerPackDstDir is the config key for knowing where to place the files
	// to be packed with the bundler.
	KeyBundlerPackDstDir = "bundler.pack.dst"
	// KeyBundlerPackExtraDirs is the config key for knowing from where the files are
	// picked to be packed with the bundler.
	KeyBundlerPackExtraDirs = "bundler.pack.srcs"
)

// ExecuteBundlerOperations parses the command line and environment
// to find operations that the bundler should execute: collect migrations,
// collect static files, run migrations, etc..
// This function is a helper to be able to run those operations from
// the same server executable file.
func ExecuteBundlerOperations(v *viper.Viper, l logs.Logger, confPrefix string) {
	shouldCollect := v.GetBool(confPrefix + KeyBundlerCollectMigrations)
	if shouldCollect {
		if err := UpdateMigrationsFromConfig(v, l, confPrefix); err != nil {
			l.Err(err, fmt.Sprintf("cannot update migration from config: %s", err.Error()))
		}
	}

	migrateVer := v.GetString(confPrefix + KeyBundlerMigrate)
	if len(migrateVer) > 0 {
		err := ApplyMigrationsFromConfig(migrateVer, v, l, confPrefix)
		if err != nil {
			l.Err(err, fmt.Sprintf("cannot apply migration: %s", err.Error()))
		}
	}

	bundleDstDir := v.GetString(confPrefix + KeyBundlerPackDstDir)
	if len(bundleDstDir) > 0 {
		scanDirs := v.GetStringSlice(confPrefix + KeyBundlerPackExtraDirs)
		projDir, err := os.Getwd()
		if err == nil {
            // TODO: we might have a conf variant set! we must use
            // the hardcoded "prod" only as a fallback
			err = PrepareBundle(projDir, bundleDstDir, scanDirs, "prod")
		}
		if err != nil {
			l.Err(err, fmt.Sprintf("cannot prepare bundle: %s", err.Error()))
			return
		}
	} else {
		l.Info("not bundling enabled")
	}
}

// UpdateMigrationsFromConfig reads the configuration values that are set to
// know the directory from where to collect migrations, and the directory to
// put the newly found migrations, to call the [UpdateMigrations] function that
// performs the actual collection.
func UpdateMigrationsFromConfig(v *viper.Viper, l logs.Logger, confPrefix string) error {
	dstDir := v.GetString(confPrefix + KeyBundlerCollectMigrationsDstDir)
	scanDirs := v.GetStringSlice(confPrefix + KeyBundlerCollectMigrationsScanDirs)
	return UpdateMigrations(dstDir, scanDirs, l)
}

// ApplyMigrationsFromConfig reads the migrations firectory from config and
// applies them.
func ApplyMigrationsFromConfig(migrateVer string, v *viper.Viper,
	l logs.Logger, confPrefix string) error {

	dstDir := v.GetString(confPrefix + KeyBundlerCollectMigrationsDstDir)
	absDir, err := filepath.Abs(dstDir)
	if err != nil {
		return fmt.Errorf("cannot find abs dir %s: %s", dstDir, err.Error())
	}
	fURL := fmt.Sprintf("file://%s", absDir)

	dbConf, err := config.ReadSQLDBConfig(confPrefix)
	if err != nil {
		return fmt.Errorf("cannot read sql db config: %s", err.Error())
	}
	pgURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		dbConf.User, dbConf.Password, dbConf.Host, dbConf.Port, dbConf.Name)
	mig, err := migrate.New(fURL, pgURL)
	if err != nil {
		return fmt.Errorf("cannot apply %s migration to %s: %s",
			fURL, pgURL, err.Error())
	}

	if migrateVer == "up" {
		if err := mig.Up(); err != nil && err != migrate.ErrNoChange {
			return err
		}
	} else if migrateVer == "down" {
		if err := mig.Down(); err != nil {
			return err
		}
	} else {
		// read the value as a number, we run ALWAYS on 64 bit,
		// because migrations numbers are that long
		verNum64, err := strconv.ParseUint(migrateVer, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot read migration version %s: %s",
				migrateVer, err.Error())
		}
		verNumber := uint(verNum64)
		if err := mig.Migrate(verNumber); err != nil {
			return err
		}
	}
	return nil
}
