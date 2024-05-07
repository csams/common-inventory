package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/csams/common-inventory/cmd/migrate"
	"github.com/csams/common-inventory/cmd/serve"

	"github.com/csams/common-inventory/pkg/server"
	"github.com/csams/common-inventory/pkg/storage"
)

var (
	version     = "0.1.0"
	programName = "common-inventory"
	cfgFile     string

	logLevel   = new(slog.LevelVar) // Info by default
	logOptions = &slog.HandlerOptions{Level: logLevel}
	rootLog    = slog.New(slog.NewJSONHandler(os.Stderr, logOptions))

	rootCmd = &cobra.Command{
		Use:     programName,
		Version: version,
		Short:   "A simple common inventory system",
	}

	options = struct {
		Storage *storage.Options `mapstructure:"storage"`
		Server  *server.Options  `mapstructure:"server"`
	}{
		storage.NewOptions(),
		server.NewOptions(),
	}
)

// Execute is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		rootLog.Error(err.Error())
		os.Exit(1)
	}
}

// init adds all child commands to the root command and sets flags appropriately.
func init() {
	// initializers are run as part of Command.PreRun
	cobra.OnInitialize(initConfig)

	configHelp := fmt.Sprintf("config file (default is $PWD/.%s.yaml)", programName)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", configHelp)
	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))

	migrateCmd := migrate.NewCommand(options.Storage, rootLog.WithGroup("storage"))
	rootCmd.AddCommand(migrateCmd)
	viper.BindPFlags(migrateCmd.Flags())

	serveCmd := serve.NewCommand(options.Server, options.Storage, rootLog.WithGroup("server"))
	rootCmd.AddCommand(serveCmd)
	viper.BindPFlags(serveCmd.Flags())
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(".")
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")

		configName := fmt.Sprintf(".%s", programName)
		viper.SetConfigName(configName)
	}

	viper.SetEnvPrefix(programName)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		msg := fmt.Sprintf("Using config file: %s", viper.ConfigFileUsed())
		rootLog.Debug(msg)
	}

	// put the values into the options struct.
	viper.Unmarshal(&options)
}
