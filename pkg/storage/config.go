package storage

import (
	"github.com/csams/common-inventory/pkg/storage/postgres"
	"github.com/csams/common-inventory/pkg/storage/sqlite3"
)

type Config struct {
	Database string
	DSN      string

	Postgres *postgres.Config
	SqlLite3 *sqlite3.Config
}

type completedConfig struct {
	Database string
	DSN      string
}

type CompletedConfig struct {
	*completedConfig
}

func NewConfig(o *Options) *Config {
	cfg := &Config{
		Database: o.Database,
	}

	switch o.Database {
	case "postgres":
		cfg.Postgres = postgres.NewConfig(o.Postgres)
	case "sqlite3":
		cfg.SqlLite3 = sqlite3.NewConfig(o.SqlLite3)
	}

	return cfg
}

func (c *Config) Complete() (CompletedConfig, error) {
	cfg := &completedConfig{
		Database: c.Database,
		DSN:      c.DSN,
	}

	if c.DSN != "" {
		return CompletedConfig{cfg}, nil
	}

	switch c.Database {
	case "postgres":
		c, err := c.Postgres.Complete()
		if err != nil {
			return CompletedConfig{}, err
		}
		cfg.DSN = c.DSN
	case "sqlite3":
		cfg.DSN = c.SqlLite3.Complete().DSN
	}

	return CompletedConfig{cfg}, nil
}
