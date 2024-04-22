package storage

import (
	"errors"

	"github.com/csams/common-inventory/pkg/storage/postgres"
	"github.com/csams/common-inventory/pkg/storage/sqlite3"
	"github.com/spf13/pflag"
)

type Options struct {
	Postgres *postgres.Options `mapstructure:"postgres"`
	SqlLite3 *sqlite3.Options  `mapstructure:"sqlite3"`
	Database string            `mapstructure:"database"`
}

func NewOptions() *Options {
	return &Options{
		Postgres: postgres.NewOptions(),
		SqlLite3: sqlite3.NewOptions(),
		Database: "sqlite3",
	}
}

func (o *Options) AddFlags(fs *pflag.FlagSet, prefix string) {
	if prefix != "" {
		prefix = prefix + "."
	}

	fs.String(prefix+"database", o.Database, "The database type to use.  Either sqlite3 or postgres.")

	o.Postgres.AddFlags(fs, prefix+"postgres")
	o.SqlLite3.AddFlags(fs, prefix+"sqlite3")
}

func (o *Options) Complete() error {
	return nil
}

func (o *Options) Validate() []error {
	var errs []error
	if o.Database != "postgres" && o.Database != "sqlite3" {
		errs = append(errs, errors.New("database must be either postgres or sqlite3"))
	}

	switch o.Database {
	case "postgres":
		errs = append(errs, o.Postgres.Validate()...)
	case "sqlite3":
		errs = append(errs, o.SqlLite3.Validate()...)
	}

	return errs
}
