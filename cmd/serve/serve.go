package serve

import (
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/csams/common-inventory/pkg/controllers"
	"github.com/csams/common-inventory/pkg/errors"
	"github.com/csams/common-inventory/pkg/server"
	"github.com/csams/common-inventory/pkg/storage"
)

func NewCommand(serverOptions *server.Options, storageOptions *storage.Options, log *slog.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the inventory server",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := storageOptions.Complete(); err != nil {
				return err
			}

			if errs := storageOptions.Validate(); errs != nil {
				return errors.NewAggregate(errs)
			}

			if err := serverOptions.Complete(); err != nil {
				return err
			}

			if errs := serverOptions.Validate(); errs != nil {
				return errors.NewAggregate(errs)
			}

			storageConfig, err := storage.NewConfig(storageOptions).Complete()
			if err != nil {
				return err
			}

			serverConfig := server.NewConfig(serverOptions).Complete()

			db, err := storage.New(storageConfig)
			if err != nil {
				return err
			}

			handler := controllers.NewHandler(db, log)
			server, err := server.New(serverConfig, handler, log)
			if err != nil {
				return err
			}

			return server.PrepareRun().Run()
		},
	}

	serverOptions.AddFlags(cmd.Flags(), "server")
	storageOptions.AddFlags(cmd.Flags(), "storage")

	return cmd
}
