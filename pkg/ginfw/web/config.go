package web

import (
	"fmt"

	"github.com/dhontecillas/hfw/pkg/config"
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
	Port             int    `json:"port"`
	Host             string `json:"host"`
	HTMLTemplatesDir string `json:"htmltemplates"`
	ServeStatic      bool   `json:"servestatic"`
	StaticDir        string `json:"staticdir"`
}

func (c *Config) Validate() error {
	if c.Port <= 0 {
		c.Port = 7777
	}
	if c.Port > 65536 {
		return fmt.Errorf("bad port number")
	}
	if c.Host == "" {
		c.Host = "127.0.0.1"
	}
	if c.HTMLTemplatesDir == "" {
		c.HTMLTemplatesDir = "./data/HTML_templates"
	}
	if c.StaticDir == "" {
		c.StaticDir = "./data/static"
	}
	return nil
}

// NewDefaultConfig creates the default configuration
// for a service to run.string
func NewDefaultConfig(cldr config.ConfLoader) *Config {
	onErrConf := &Config{
		Port:             7777,
		Host:             "127.0.0.1",
		HTMLTemplatesDir: "./data/HTML_templates",
		ServeStatic:      false,
		StaticDir:        "./data/static",
	}
	var conf Config
	err := cldr.Parse(&conf)
	if err != nil {
		return onErrConf
	}
	err = conf.Validate()
	if err != nil {
		return onErrConf
	}
	return &conf
}

// Address returns the address of the web service
func (c *Config) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
