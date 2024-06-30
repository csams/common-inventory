package server

import (
	"crypto/tls"
	"crypto/x509"
	"io"
	"os"
)

type Config struct {
	Options   *Options
	TLSConfig *tls.Config
}

type completedConfig struct {
	Options       *Options
	SecureServing bool
	TLSConfig     *tls.Config
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

func (c *Config) getTSLConfig() (*tls.Config, error) {
	if c.TLSConfig == nil {
		if c.Options.CertOpt > int(tls.NoClientCert) {
			var caCertPool *x509.CertPool
			if file, err := os.Open(c.Options.ClientCAFile); err == nil {
				if caCert, err := io.ReadAll(file); err == nil {
					caCertPool = x509.NewCertPool()
					caCertPool.AppendCertsFromPEM(caCert)
				} else {
					return nil, err
				}
			}

			c.TLSConfig = &tls.Config{
                ServerName: c.Options.SNI,
                ClientAuth: tls.ClientAuthType(c.Options.CertOpt),
                ClientCAs: caCertPool,
                MinVersion: tls.VersionTLS12,
            }
		}
	}

	return c.TLSConfig, nil
}

func (c *Config) Complete() (CompletedConfig, error) {
	tlsConfig, err := c.getTSLConfig()

	return CompletedConfig{&completedConfig{
		Options:       c.Options,
		SecureServing: c.Options.ServingCertFile != "" || c.Options.PrivateKeyFile != "" || c.Options.CertOpt > int(tls.NoClientCert),
		TLSConfig:     tlsConfig,
	}}, err
}
