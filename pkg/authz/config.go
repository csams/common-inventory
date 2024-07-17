package authz

import (
	"context"

	"github.com/csams/common-inventory/pkg/authz/kessel"
)

type Config struct {
	Authz  string
	Kessel kessel.Config
}

func NewConfig(o *Options) *Config {
	return &Config{
		Authz:  o.Authz,
		Kessel: *kessel.NewConfig(o.Kessel),
	}
}

type completedConfig struct {
	Authz  string
	Kessel kessel.CompletedConfig
}

type CompletedConfig struct {
	*completedConfig
}

func (c *Config) Complete(ctx context.Context) (CompletedConfig, []error) {
	cfg := &completedConfig{}

	if c.Authz == "kessel" {
		if ksl, errs := c.Kessel.Complete(ctx); errs != nil {
			return CompletedConfig{}, nil
		} else {
			cfg.Kessel = ksl
		}
	}

	return CompletedConfig{cfg}, nil
}
