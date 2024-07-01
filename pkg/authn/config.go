package authn

import (
	"github.com/csams/common-inventory/pkg/authn/oidc"
	"github.com/csams/common-inventory/pkg/authn/psk"
)

type Config struct {
	Oidc          *oidc.Config
	PreSharedKeys *psk.Config
}

func NewConfig(o *Options) *Config {
	cfg := &Config{}

	if len(o.Oidc.AuthorizationServerURL) > 0 {
		cfg.Oidc = oidc.NewConfig(o.Oidc)
	}

	if len(o.PreSharedKeys.PreSharedKeyFile) > 0 {
		cfg.PreSharedKeys = psk.NewConfig(o.PreSharedKeys)

	}

	return cfg
}

type completedConfig struct {
	Oidc          *oidc.CompletedConfig
	PreSharedKeys *psk.CompletedConfig
}

type CompletedConfig struct {
	*completedConfig
}

func (c *Config) Complete() (CompletedConfig, []error) {
	var errs []error
	cfg := CompletedConfig{&completedConfig{}}

	if c.Oidc != nil {
		if o, err := c.Oidc.Complete(); err == nil {
			cfg.Oidc = &o
		} else {
			errs = append(errs, err)
		}
	}

	if c.PreSharedKeys != nil {
		if o, err := c.PreSharedKeys.Complete(); err == nil {
			cfg.PreSharedKeys = &o
		} else {
			errs = append(errs, err)
		}
	}

	if errs != nil {
		return CompletedConfig{completedConfig: &completedConfig{}}, errs
	}

	return cfg, nil
}
