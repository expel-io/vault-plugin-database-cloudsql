package main

import (
	"context"
	"fmt"
	"os"

	"github.com/expel-io/vault-plugin-database-cloudsql/cloudsql"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/hashicorp/vault/api"
	dbplugin "github.com/hashicorp/vault/sdk/database/dbplugin/v5"
	"github.com/pkg/errors"
)

// Serve launches the plugin
// testServerChan is optional, if passed it will configure the plugin server in test mode.
// if the server is in test mode it will return the ReattachConfig through the channel,
// and it can then be used to terminate the test server.
// See: https://github.com/hashicorp/go-plugin/blob/v1.4.4/server.go#L117
func Serve(ctx context.Context, testServerChan chan *plugin.ReattachConfig) {
	apiClientMeta := &api.PluginAPIClientMeta{}
	flags := apiClientMeta.FlagSet()
	var flagDBType string
	var flagLogLevel string
	var flagMultiplex bool
	flags.StringVar(&flagDBType, "db-type", cloudsql.Postgres.String(), "can be: 'cloudsql-postgres'")
	flags.StringVar(&flagLogLevel, "log-level", "info", "can be: 'trace', 'debug', 'info', 'warn', 'error', 'off'")
	flags.BoolVar(&flagMultiplex, "multiplex", true, "Whether to enable plugin multiplexing. can be: 'true' or 'false'")
	err := flags.Parse(os.Args[1:])
	if err != nil {
		fmt.Printf("unable to parse plugin arguments: %s", err)
		os.Exit(1)
	}

	logger := hclog.New(&hclog.LoggerOptions{
		Name:  "vault-plugin-database-cloudsql",
		Level: hclog.LevelFromString(flagLogLevel),
	})

	// provide factory to initialize a "cloudsql" database plugin
	pluginFactory := func() (interface{}, error) {
		cloudsqlDatabase, err := cloudsql.New(cloudsql.DBType(flagDBType))
		if err != nil {
			return nil, errors.Wrap(err, "failed to get new instance of cloudsql database plugin")
		}
		return cloudsqlDatabase, nil
	}

	serveConfig, err := initServeConfig(flagMultiplex, pluginFactory, logger)
	if err != nil {
		logger.Error("failed to initialize database plugin. aborting now.", err)
		os.Exit(1)
	}
	if testServerChan != nil {
		// if running in test mode, use channel to pry into the plugin's lifecycle
		serveConfig.Test = &plugin.ServeTestConfig{
			Context:          ctx,
			ReattachConfigCh: testServerChan,
			CloseCh:          nil,
		}
		serveConfig.Logger = hclog.NewNullLogger()
	} else {
		serveConfig.Logger = logger
	}
	// Vault communicates to plugins over RPC
	// start RPC server to which vault will connect to
	// See: https://www.vaultproject.io/docs/secrets/databases/custom#serving-a-plugin
	plugin.Serve(serveConfig)
}

func initServeConfig(flagMultiplex bool, pluginFactory dbplugin.Factory, logger hclog.Logger) (*plugin.ServeConfig, error) {
	logger.Debug("initializing cloudsql plugin with multiplexing=%t", flagMultiplex)
	var serveConfig *plugin.ServeConfig
	if flagMultiplex {
		// See: https://www.vaultproject.io/docs/plugins/plugin-architecture#plugin-multiplexing
		serveConfig = dbplugin.ServeConfigMultiplex(pluginFactory)
	} else {
		dbPlugin, err := pluginFactory()
		if err != nil {
			return nil, errors.Wrap(err, "failed to create new plugin instance")
		}
		serveConfig = dbplugin.ServeConfig(dbPlugin.(dbplugin.Database))
	}
	if serveConfig == nil {
		return nil, errors.New("failed to initialize server config for plugin")
	}
	return serveConfig, nil
}
