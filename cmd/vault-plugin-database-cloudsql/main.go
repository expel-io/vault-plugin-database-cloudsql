package main

import (
	"log"
	"os"

	"github.com/expel-io/vault-plugin-database-cloudsql/cloudsql"
	dbplugin "github.com/hashicorp/vault/sdk/database/dbplugin/v5"
)

func main() {
	// TODO document this
	dbType, err := cloudsql.New()
	if err != nil {
		// TODO improve error handling
		log.Println(err)
		os.Exit(1)
	}

	// TODO what is this? what does it do?
	dbplugin.Serve(dbType.(dbplugin.Database))
}
