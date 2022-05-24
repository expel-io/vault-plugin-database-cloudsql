package main

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/expel-io/vault-plugin-database-cloudsql/cloudsql"
	"github.com/hashicorp/go-plugin"
)

func TestServe(t *testing.T) {
	os.Args = []string{"my-plugin"}
	ctx := context.Background()
	reattachConfigCh := make(chan *plugin.ReattachConfig)

	// launch the plugin server in the background
	go Serve(ctx, reattachConfigCh)

	// wait for the plugin to read the config and launch
	select {
	case <-reattachConfigCh:
	case <-time.After(time.Second * 10):
		t.Fatal("timed out waiting for plugin to launch")
	}

	// assert that the driver was registered correctly
	foundDriver := false
	for _, v := range sql.Drivers() {
		if v == cloudsql.Postgres.String() {
			foundDriver = true
		}
	}
	if !foundDriver {
		t.Error("expected the driver 'cloudsql-postgres' to be registered but was not found")
	}
}
