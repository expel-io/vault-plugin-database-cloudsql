package cloudsql

import (
	"database/sql"
	"reflect"
	"testing"
	"unsafe"

	"github.com/hashicorp/vault/plugins/database/postgresql"
	dbplugin "github.com/hashicorp/vault/sdk/database/dbplugin/v5"
)

func TestNewDelegatesToVaultPostgresPlugin(t *testing.T) {
	// initialize for a postgres database
	dbPlugin, err := New(Postgres)
	if err != nil {
		t.Error("failed to initialize CloudSQL database", err)
		return
	}

	wrappedDBMiddleware := dbPlugin.(dbplugin.DatabaseErrorSanitizerMiddleware)
	// get the CloudSQL db plugin instance from the wrapping middleware
	wrappedPlugin := reflect.ValueOf(&wrappedDBMiddleware).Elem().FieldByName("next")
	safePointerToPlugin := reflect.NewAt(wrappedPlugin.Type(), unsafe.Pointer(wrappedPlugin.UnsafeAddr())).Elem()
	cloudSQLInterface := safePointerToPlugin.Interface()
	cloudSQLPlugin := cloudSQLInterface.(*CloudSQL)

	// assert that the correct delegateVaultPlugin was initialized
	_, ok := cloudSQLPlugin.delegateVaultPlugin.(*postgresql.PostgreSQL)
	if !ok {
		t.Errorf("expected type of delegated database vault plugin to be of type '*postgresql.PostgreSQL' but got '%s'", reflect.TypeOf(cloudSQLPlugin.delegateVaultPlugin))
	}

	// assert that the driver was registered correctly
	foundDriver := false
	for _, v := range sql.Drivers() {
		if v == Postgres.String() {
			foundDriver = true
		}
	}
	if !foundDriver {
		t.Error("expected the driver 'cloudsql-postgres' to be registered but was not found")
	}
}
