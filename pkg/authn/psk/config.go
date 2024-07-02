package psk

import (
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	PreSharedKeyFile string
	Keys             IdentityMap
}

type completedConfig struct {
	Keys IdentityMap
}

type CompletedConfig struct {
	*completedConfig
}

func NewConfig(o *Options) *Config {
	return &Config{
		PreSharedKeyFile: o.PreSharedKeyFile,
	}
}

func (c *Config) Complete() (CompletedConfig, error) {
	if len(c.Keys) == 0 {
		if err := c.loadPreSharedKeys(); err != nil {
			return CompletedConfig{}, err
		}
	}

	return CompletedConfig{&completedConfig{
		Keys: c.Keys,
	}}, nil
}

func (c *Config) loadPreSharedKeys() error {
	// TODO: fsnotify to reload the keys when the PSK file changes.
	if file, err := os.Open(c.PreSharedKeyFile); err == nil {
		defer file.Close()
		data, err := io.ReadAll(file)
		if err == nil {
			if err := yaml.Unmarshal(data, c.Keys); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		return err
	}
	return nil
}
