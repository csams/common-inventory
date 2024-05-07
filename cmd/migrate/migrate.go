package migrate

import (
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/csams/common-inventory/pkg/errors"
	"github.com/csams/common-inventory/pkg/models"
	"github.com/csams/common-inventory/pkg/storage"
)

func NewCommand(options *storage.Options, log *slog.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Create or migrate the database tables",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := options.Complete(); err != nil {
				return err
			}

			if errs := options.Validate(); errs != nil {
				return errors.NewAggregate(errs)
			}

			config := storage.NewConfig(options).Complete()

			db, err := storage.New(config)
			if err != nil {
				return err
			}

			return models.Migrate(db)
		},
	}

	options.AddFlags(cmd.Flags(), "storage")

	return cmd
}
