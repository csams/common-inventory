package server

import (
	"fmt"

	"github.com/spf13/pflag"
)

type Options struct {
	Address  string `mapstructure:"address"`
	CertFile string `mapstructure:"certfile"`
	KeyFile  string `mapstructure:"keyfile"`
}

func NewOptions() *Options {
	return &Options{
		Address: "localhost:9090",
	}
}

func (o *Options) AddFlags(fs *pflag.FlagSet, prefix string) {
	if prefix != "" {
		prefix = prefix + "."
	}

	fs.StringVar(&o.Address, prefix+"address", o.Address, "the host and port on which to listen")
	fs.StringVar(&o.CertFile, prefix+"certfile", o.CertFile, "the file containing the server's serving certificate")
	fs.StringVar(&o.KeyFile, prefix+"keyfile", o.KeyFile, "the file containing the server's private key for the serving cert")
}

func (o *Options) Complete() error {
	return nil
}

func (o *Options) Validate() []error {
	var errors []error
	if (o.CertFile == "" && o.KeyFile != "") || (o.CertFile != "" && o.KeyFile == "") {
		err := fmt.Errorf("Both certfile and keyfile must be populated")
		errors = append(errors, err)
	}
	return errors
}
