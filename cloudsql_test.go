package cloudsql

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/hashicorp/vault/plugins/database/postgresql"
	dbplugin "github.com/hashicorp/vault/sdk/database/dbplugin/v5"
)

func TestNewDelegatesToVaultPostgresPlugin(t *testing.T) {
	// initialize for a postgres database
	dbPlugin, err := New("cloudsql-postgres")
	if err != nil {
		t.Error("failed to initialize cloudsql database", err)
		return
	}

	wrappedDBMiddleware := dbPlugin.(dbplugin.DatabaseErrorSanitizerMiddleware)
	// Get the cloudsql db plugin instance from the wrapping middleware
	wrappedCloudsqlDB := reflect.ValueOf(&wrappedDBMiddleware).Elem().FieldByName("next")
	safePointerToCloudsqlDB := reflect.NewAt(wrappedCloudsqlDB.Type(), unsafe.Pointer(wrappedCloudsqlDB.UnsafeAddr())).Elem()
	cloudsqlInterface := safePointerToCloudsqlDB.Interface()
	cloudsqlDB := cloudsqlInterface.(*CloudSQL)

	// assert that the correct targetDBDelegate was initialized
	_, ok := cloudsqlDB.targetDBDelegate.(*postgresql.PostgreSQL)
	if !ok {
		t.Errorf("expected type of targetDBDelegate to be of type *postgresql.PostgreSQL but got %s\n", reflect.TypeOf(cloudsqlDB.targetDBDelegate))
	}
	// Todo assert that the driver was registered correctly
}
