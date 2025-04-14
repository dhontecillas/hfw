package config

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/dhontecillas/hfw/pkg/extdeps"
	"github.com/dhontecillas/hfw/pkg/obs"
)

const (
	// KeyConfVariant is the environment variable to select
	// a different configuration file. If not set the default
	// config file is: config.yaml, but if set, the file would
	// be config.[CONFVARIANT].yaml.
	KeyConfVariant string = "CONFVARIANT"
	// DefaultConfigPath is the default directory from where to
	// load a configuration file.
	DefaultConfigPath = "./config"
	// BundleConfigPath is a fallback config directory to load
	// configuration files from.
	BundleConfigPath = "./bundle/config"
)

// InitConfig loads the global MapConf instance, to look for
// some files, and use environment vars (changing '_' to '.')
// It looks for a config file config.json, in
// a couple of default directories ("./config", "./bundle/config").
// An alternative config file can be selected using an environment
// var: `CONFVARIANT` or `[config_prefix]CONFVARIANT`, in which
// case the file used will be `config.[CONFCARIANT value].yaml`.
func InitConfig(confPrefix string) (ConfLoader, error) {
	envPathSep := "_"
	confVariantKey := KeyConfVariant
	if len(confPrefix) > 0 {
		confVariantKey = strings.ToUpper(confPrefix) + envPathSep + confVariantKey
	}
	confFile := "config.json"
	confVariant := os.Getenv(confVariantKey)
	if len(confVariant) > 0 {
		confFile = "config." + confVariant + ".json"
	}

	configFilePaths := []string{
		DefaultConfigPath,
		BundleConfigPath,
		".",
	}

	conf := newMapConf(nil)
	var content []byte
	var err error
	var fullPathConfFile string
	for _, v := range configFilePaths {
		fullPathConfFile = path.Join(v, confFile)
		content, err = os.ReadFile(fullPathConfFile)
		if err == nil {
			// TODO: log the name of the config loaded
			// check that the logger is already set up
			break
		}
	}
	if len(content) == 0 {
		fmt.Printf("Cannot find config file.\n")
	} else {
		mcJSON, err := newMapConfFromJSON(content)
		if err != nil {
			fmt.Printf("Cannot parse json file from: %s\n", fullPathConfFile)
		}
		conf.Merge(mcJSON)
	}

	mcEnv := newMapConfFromEnv(confPrefix, envPathSep)
	conf.Merge(mcEnv)
	return conf, nil
}

// BuildExternalServices creates an external services instance based
// on configuration.
func BuildExternalServices(cldr ConfLoader,
	insBuilderFn obs.InsighterBuilderFn,
	insFlush func()) *extdeps.ExternalServicesBuilder {

	ins := insBuilderFn()

	mailConf, err := ReadMailerConfig(ins, cldr)
	if err != nil {
		panic(fmt.Sprintf("cannot read mailer configuration: %s", err.Error()))
	}
	mailer, err := CreateMailer(ins, mailConf)
	if err != nil {
		panic(fmt.Sprintf("cannot configure mailer: %s", err.Error()))
	}

	dbConf, err := ReadSQLDBConfig(cldr)
	if err != nil {
		panic(fmt.Sprintf("cannot read sql db config: %s", err.Error()))
	}
	sql := CreateSQLDB(ins, dbConf)
	if sql == nil {
		panic("cannot create sql db connection")
	}

	notificationsConf, err := ReadNotificationsConfig(ins, cldr)
	if err != nil {
		panic("cannot read notifications config")
	}
	composer, err := CreateNotificationsComposer(ins, notificationsConf, mailer)
	if err != nil {
		panic("cannot create notifications")
	}
	return extdeps.NewExternalServicesBuilder(
		insBuilderFn, insFlush,
		mailer,
		sql,
		composer)
}
