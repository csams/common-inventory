package server

type Config struct {
	Options *Options
}

type completedConfig struct {
	Options       *Options
	SecureServing bool
}

// CompletedConfig can be constructed only from Config.Complete
type CompletedConfig struct {
	*completedConfig
}

func NewConfig(o *Options) *Config {
	return &Config{
		Options: o,
	}
}

func (c *Config) Complete() CompletedConfig {
	return CompletedConfig{&completedConfig{
		Options:       c.Options,
		SecureServing: c.Options.CertFile != "" && c.Options.KeyFile != "",
	}}
}
