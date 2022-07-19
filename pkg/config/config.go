package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"

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

// InitConfig configures the global viper instance, to look for
// some files, and use environment vars (changing '_' to '.')
// It looks for a config file (config.yaml, config.json...), in
// a couple of default directories ("./config", "./bundle/config").
// An alternative config file can be selected using an environment
// var: `CONFVARIANT` or `[config_prefix]CONFVARIANT`, in which
// case the file used will be `config.[CONFCARIANT value].yaml`.
func InitConfig(confPrefix string) error {
	viper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	confVariantKey := KeyConfVariant
	if len(confPrefix) > 0 {
		confVariantKey = replacer.Replace(strings.ToUpper(confPrefix)) + confVariantKey
	}
	confVariant := viper.GetString(confVariantKey) // os.Getenv(confVariantKey)
	if len(confVariant) > 0 {
		viper.SetConfigName(fmt.Sprintf("config.%s", confVariant))
	} else {
		viper.SetConfigName("config")
	}
	viper.AddConfigPath(DefaultConfigPath)
	viper.AddConfigPath(BundleConfigPath)

	err := viper.ReadInConfig()
	if err != nil {
		return err
	}
	return nil
}

// BuildExternalServices creates an external services instance based
// on configuration.
func BuildExternalServices(confPrefix string,
	insBuilderFn obs.InsighterBuilderFn,
	insFlush func()) *extdeps.ExternalServices {

	ins := insBuilderFn()

	mailConf, err := ReadMailerConfig(ins, confPrefix)
	if err != nil {
		panic(fmt.Sprintf("cannot read mailer configuration: %s", err.Error()))
	}
	mailer, err := CreateMailer(ins, mailConf)
	if err != nil {
		panic(fmt.Sprintf("cannot configure mailer: %s", err.Error()))
	}

	dbConf, err := ReadSQLDBConfig(confPrefix)
	if err != nil {
		panic(fmt.Sprintf("cannot read sql db config: %s", err.Error()))
	}
	sql := CreateSQLDB(ins, dbConf)
	if sql == nil {
		panic("cannot create sql db connection")
	}

	notificationsConf, err := ReadNotificationsConfig(ins, confPrefix)
	if err != nil {
		panic("cannot read notifications config")
	}
	notifier, err := CreateNotifications(ins, notificationsConf, mailer)
	if err != nil {
		panic("cannot create notifications")
	}
	return extdeps.NewExternalServices(
		insBuilderFn, insFlush,
		mailer,
		sql,
		notifier)
}
