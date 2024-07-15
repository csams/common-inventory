package serve

import (
	"context"
	"fmt"
	"log/slog"

	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/csams/common-inventory/pkg/authn"
	"github.com/csams/common-inventory/pkg/controllers"
	"github.com/csams/common-inventory/pkg/errors"
	"github.com/csams/common-inventory/pkg/eventing"
	eventingapi "github.com/csams/common-inventory/pkg/eventing/api"
	"github.com/csams/common-inventory/pkg/server"
	"github.com/csams/common-inventory/pkg/storage"
)

func NewCommand(
	serverOptions *server.Options,
	storageOptions *storage.Options,
	authnOptions *authn.Options,
	eventingOptions *eventing.Options,
	log *slog.Logger,
) *cobra.Command {
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

			// configure authn
			if errs := eventingOptions.Complete(); errs != nil {
				return errors.NewAggregate(errs)
			}

			if errs := eventingOptions.Validate(); errs != nil {
				return errors.NewAggregate(errs)
			}

			eventingConfig, errs := eventing.NewConfig(eventingOptions).Complete()
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

			eventingManager, err := eventing.New(eventingConfig, log)

			// bring up the server
			rootHandler := controllers.NewRootHandler(db, authenticator, eventingManager, log)
			server := server.New(serverConfig, rootHandler, log)
			if err != nil {
				return err
			}

			srvErrs := make(chan error)
			go func() {
				srvErrs <- server.Run()
			}()

			quit := make(chan os.Signal, 1)
			signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

			shutdown := gracefulShutdown(server, eventingManager, log)

			select {
			case err := <-srvErrs:
				shutdown(err)
			case sig := <-quit:
				shutdown(sig)
			case emErr := <-eventingManager.Errs():
				shutdown(emErr)
			}
			return nil
		},
	}

	serverOptions.AddFlags(cmd.Flags(), "server")
	storageOptions.AddFlags(cmd.Flags(), "storage")
	authnOptions.AddFlags(cmd.Flags(), "authn")
	eventingOptions.AddFlags(cmd.Flags(), "eventing")

	return cmd
}

func gracefulShutdown(srv *server.Server, em eventingapi.Manager, log *slog.Logger) func(reason interface{}) {
	return func(reason interface{}) {
		log.Info(fmt.Sprintf("Server Shutdown: %s", reason))

		timeout := srv.HttpServer.ReadTimeout
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Error(fmt.Sprintf("Error Gracefully Shutting Down API: %v", err))
		}

		ctx, cancel = context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := em.Shutdown(ctx); err != nil {
			log.Error(fmt.Sprintf("Error Gracefully Shutting Down Eventing: %v", err))
		}
	}
}
