package serve

import (
	"log/slog"

	"github.com/spf13/cobra"

    "github.com/csams/common-inventory/pkg/authn"
	"github.com/csams/common-inventory/pkg/controllers"
	"github.com/csams/common-inventory/pkg/errors"
	"github.com/csams/common-inventory/pkg/server"
	"github.com/csams/common-inventory/pkg/storage"
)

func NewCommand(serverOptions *server.Options, storageOptions *storage.Options, authnOptions *authn.Options, log *slog.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the inventory server",
		RunE: func(cmd *cobra.Command, args []string) error {
			// configure storage
			if errs := storageOptions.Complete(); errs != nil {
				return errors.NewAggregate(errs)
			}

			if errs := storageOptions.Validate(); errs != nil {
				return errors.NewAggregate(errs)
			}

			storageConfig := storage.NewConfig(storageOptions).Complete()

            // configure authn
			if errs := authnOptions.Complete(); errs != nil {
				return errors.NewAggregate(errs)
			}

			if errs := serverOptions.Validate(); errs != nil {
				return errors.NewAggregate(errs)
			}

			authnConfig, errs := authn.NewConfig(authnOptions).Complete()
			if errs != nil {
				return errors.NewAggregate(errs)
			}

			// configure the server
			if errs := serverOptions.Complete(); errs != nil {
				return errors.NewAggregate(errs)
			}

			if errs := serverOptions.Validate(); errs != nil {
				return errors.NewAggregate(errs)
			}

			serverConfig, err := server.NewConfig(serverOptions).Complete()
			if err != nil {
				return err
			}

			// bring up storage
			db, err := storage.New(storageConfig)
			if err != nil {
				return err
			}

            // bring up the authenticator
            authenticator, err := authn.New(authnConfig)
            if err != nil {
                return err
            }

			// bring up the server
			rootHandler := controllers.NewRootHandler(db, authenticator, log)
			server := server.New(serverConfig, rootHandler, log)
			if err != nil {
				return err
			}

			return server.PrepareRun().Run()
		},
	}

	serverOptions.AddFlags(cmd.Flags(), "server")
	storageOptions.AddFlags(cmd.Flags(), "storage")
    authnOptions.AddFlags(cmd.Flags(), "authn")

	return cmd
}
