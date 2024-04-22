package sqlite3

import "github.com/spf13/pflag"

type Options struct {
	DSN string `mapstructure:"dsn"`
}

func NewOptions() *Options {
	return &Options{
		DSN: "inventory.db",
	}
}

func (o *Options) AddFlags(fs *pflag.FlagSet, prefix string) {
	if prefix != "" {
		prefix = prefix + "."
	}

	fs.String(prefix+"dsn", o.DSN, "The connection string to use for sqlite3.")
}

func (o *Options) Complete() error {
	return nil
}

func (o *Options) Validate() []error {
	return nil
}
