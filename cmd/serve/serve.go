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
			// configure storage
			if err := storageOptions.Complete(); err != nil {
				return err
			}

			if errs := storageOptions.Validate(); errs != nil {
				return errors.NewAggregate(errs)
			}

			storageConfig := storage.NewConfig(storageOptions).Complete()

			// configure the server
			if err := serverOptions.Complete(); err != nil {
				return err
			}

			if errs := serverOptions.Validate(); errs != nil {
				return errors.NewAggregate(errs)
			}

			serverConfig := server.NewConfig(serverOptions).Complete()

			// bring up storage
			db, err := storage.New(storageConfig)
			if err != nil {
				return err
			}

			// bring up the server
			rootHandler := controllers.NewRootHandler(db, log)
			server, err := server.New(serverConfig, rootHandler, log)
			if err != nil {
				return err
			}

			return server.PrepareRun().Run()
		},
	}

	storageOptions.AddFlags(cmd.Flags(), "storage")
	serverOptions.AddFlags(cmd.Flags(), "server")

	return cmd
}
