package main

import (
	"os"

	"github.com/expel-io/vault-plugin-database-cloudsql/cloudsql"
	"github.com/hashicorp/go-hclog"
	dbplugin "github.com/hashicorp/vault/sdk/database/dbplugin/v5"
)

func main() {
	// TODO pass debug level as an argument via terraform
	logger := hclog.New(&hclog.LoggerOptions{
		Name:  "vault-plugin-database-cloudsql",
		Level: hclog.DefaultLevel,
	})

	// TODO get target db type from arguments/terraform and pass to plugin
	// TODO convert to an enum
	cloudsqlDbInstanceType := "cloudsql-postgres"

	// initialize "cloudsql" database plugin
	cloudsqlDatabase, err := cloudsql.New(cloudsqlDbInstanceType)
	if err != nil {
		logger.Error("failed to initialize cloudsql database plugin. aborting now.", err)
		os.Exit(1)
	}

	// Vault communicates to plugins over RPC
	// start RPC server to which vault will connect to
	// See: https://www.vaultproject.io/docs/secrets/databases/custom#serving-a-plugin
	// TODO determine whether ServeMultiplex is required
	// See: https://www.vaultproject.io/docs/plugins/plugin-architecture#plugin-multiplexing
	dbplugin.Serve(cloudsqlDatabase.(dbplugin.Database))
}
