package config

import (
	"errors"
	"fmt"
	"path/filepath"
)

var (
	ErrBundlerNoSection = errors.New("no 'bundler' section")
	ErrBundlerParse     = errors.New("cannot parse 'bundler' section")
	ErrBundlerValidate  = errors.New("cannot validate 'bundler' section")
)

type BundlerConfig struct {
	Migrations BundlerMigrationsConfig `json:"migrations"`
	Pack       BundlerPackConfig       `json:"pack"`
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

// TODO: change dst to some other name, like migrations dir?
type BundlerMigrationsConfig struct {
	Dst     string   `json:"dst"`     // where to store migration files
	Scan    []string `json:"scan"`    // where to search for migrations
	Collect bool     `json:"collect"` // if we have to collect migration files
	Migrate string   `json:"migrate"`
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

// TODO: unify and use either "sources" or "scan" for the same semantic
// of places to collect files from
type BundlerPackConfig struct {
	Dst     string   `json:"dst"`
	Srcs    []string `json:"srcs"`
	Variant string   `json:"variant"`
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

func ReadBundlerConfig(cldr ConfLoader) (*BundlerConfig, error) {
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
