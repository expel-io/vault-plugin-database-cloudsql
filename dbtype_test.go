package cloudsql

import (
	"testing"
)

func TestString(t *testing.T) {
	myDBType := Postgres
	expectedDBType := "cloudsql-postgres"

	// assert that the proper string value for the enum is returned
	if myDBType.String() != expectedDBType {
		t.Errorf("expected database type to be '%s' but got '%s'", expectedDBType, myDBType.String())
	}
}

func TestFromString(t *testing.T) {
	myDBTypeString := "cloudsql-postgres"
	dbType, ok := FromString(myDBTypeString)

	// assert that the enum can be converted from a string
	if !ok {
		t.Errorf("expected 'Postgres' database type but got '%v'", dbType)
	}
}
