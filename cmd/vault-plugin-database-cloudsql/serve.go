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
	flags.StringVar(&flagDBType, "db-type", cloudsql.Postgres.String(), "can be: 'cloudsql-postgres'")
	flags.StringVar(&flagLogLevel, "log-level", "info", "can be: 'trace', 'debug', 'info', 'warn', 'error', 'off'")
	err := flags.Parse(os.Args[1:])
	if err != nil {
		fmt.Printf("unable to parse plugin arguments: %s", err)
		os.Exit(1)
	}

	logger := hclog.New(&hclog.LoggerOptions{
		Name:  "vault-plugin-database-cloudsql",
		Level: hclog.LevelFromString(flagLogLevel),
	})

	// initialize "cloudsql" database plugin
	cloudsqlDatabase, err := cloudsql.New(cloudsql.DBType(flagDBType))
	if err != nil {
		logger.Error("failed to initialize cloudsql database plugin. aborting now.", err)
		os.Exit(1)
	}

	// Vault communicates to plugins over RPC
	// start RPC server to which vault will connect to
	// See: https://www.vaultproject.io/docs/secrets/databases/custom#serving-a-plugin
	// TODO determine whether ServeMultiplex is required
	// See: https://www.vaultproject.io/docs/plugins/plugin-architecture#plugin-multiplexing
	serveConfig := dbplugin.ServeConfig(cloudsqlDatabase.(dbplugin.Database))
	if testServerChan != nil {
		serveConfig.Test = &plugin.ServeTestConfig{
			Context:          ctx,
			ReattachConfigCh: testServerChan,
			CloseCh:          nil,
		}
		serveConfig.Logger = hclog.NewNullLogger()
	} else {
		serveConfig.Logger = logger
	}

	plugin.Serve(serveConfig)
}
