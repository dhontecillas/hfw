package web

import (
	"fmt"

	"github.com/spf13/viper"
)

const (
	// KeyHost is the configuration key for the host.
	KeyHost string = "host"
	// KeyPort is the configuration key for the port.
	KeyPort string = "port"
	// KeyStaticDir is the configuration key to know where
	// static assets are stored (in case we want to serve them
	// from this application instead of using a third party
	// web server).
	KeyStaticDir string = "staticdir"
	// KeyHTMLTemplatesDir is the directory where the HTML templates
	// are stored.
	KeyHTMLTemplatesDir string = "HTMLtemplates"
)

// Config contains the information about a web
// service to run.
type Config struct {
	Port             int
	Host             string
	HTMLTemplatesDir string
	ServeStatic      bool
	StaticDir        string
}

// NewDefaultConfig creates the default configuration
// for a service to run.
func NewDefaultConfig(confPrefix string) *Config {
	conf := &Config{
		Port:             7777,
		Host:             "127.0.0.1",
		HTMLTemplatesDir: "./data/HTML_templates",
		ServeStatic:      false,
		StaticDir:        "./data/static",
	}

	staticDir := viper.GetString(confPrefix + KeyStaticDir)
	if len(staticDir) > 0 {
		conf.ServeStatic = true
		conf.StaticDir = staticDir
	}

	HTMLTemplatesDir := viper.GetString(confPrefix + KeyHTMLTemplatesDir)
	if len(HTMLTemplatesDir) > 0 {
		conf.HTMLTemplatesDir = HTMLTemplatesDir
	}

	port := viper.GetInt(confPrefix + KeyPort)
	if port > 0 && port < 65535 {
		conf.Port = port
	}

	host := viper.GetString(confPrefix + KeyHost)
	if len(host) > 0 {
		conf.Host = host
	}

	return conf
}

// Address returns the address of the web service
func (c *Config) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
