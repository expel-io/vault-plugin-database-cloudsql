package cloudsql

import (
	"context"

	"cloud.google.com/go/cloudsqlconn"
	"cloud.google.com/go/cloudsqlconn/postgres/pgxv4"
	"github.com/hashicorp/vault/plugins/database/postgresql"
	dbplugin "github.com/hashicorp/vault/sdk/database/dbplugin/v5"
	"github.com/hashicorp/vault/sdk/database/helper/connutil"
	"github.com/pkg/errors"
)

// CloudSQL implements Vault's Database interface
// See: https://www.vaultproject.io/docs/secrets/databases/custom#plugin-interface
type CloudSQL struct {
	*connutil.SQLConnectionProducer

	dbType DBType

	// delegates to the target cloudsql instance type. e.g.: postgres, mysql, etc
	delegateVaultPlugin dbplugin.Database

	connectorCleanup func() error
}

func New(dbType DBType) (interface{}, error) {
	// use the "database/sql" package for connection management.
	// this allows us to connect to the target database (postgres, mysql, etc) using
	// the common dialect provided by the "database/sql" package.
	connProducer := &connutil.SQLConnectionProducer{}
	connProducer.Type = dbType.String()

	var delegateVaultPlugin dbplugin.Database
	var connectorCleanUpFunc func() error
	var secretValuesMaskingFunc func() map[string]string
	var err error

	// determine the target cloudsql db instance type and delegate database operations to it
	if dbType == Postgres {
		delegateVaultPlugin, connectorCleanUpFunc, secretValuesMaskingFunc, err = newPostgresDatabase(dbType, connProducer)
		if err != nil {
			return nil, errors.Wrap(err, "unable to initialize connector for 'postgres' database instance")
		}
	} else {
		// no other types are supported yet
		return nil, errors.Errorf("unsupported target cloudsql database instance type: %s", dbType)
	}

	// initialize the database plugin
	cloudsqlDB := &CloudSQL{
		dbType:                dbType,
		SQLConnectionProducer: connProducer,
		delegateVaultPlugin:   delegateVaultPlugin,
		connectorCleanup:      connectorCleanUpFunc,
	}

	// Wrap the plugin with middleware to sanitize errors
	wrappedDB := dbplugin.NewDatabaseErrorSanitizerMiddleware(cloudsqlDB, secretValuesMaskingFunc)
	return wrappedDB, nil
}

// Initialize the database plugin. This is the equivalent of a constructor for the
// database object itself.
func (c *CloudSQL) Initialize(ctx context.Context, req dbplugin.InitializeRequest) (dbplugin.InitializeResponse, error) {
	return c.delegateVaultPlugin.Initialize(ctx, req)
}

// NewUser creates a new user within the database. This user is temporary in that it
// will exist until the TTL expires.
func (c *CloudSQL) NewUser(ctx context.Context, req dbplugin.NewUserRequest) (dbplugin.NewUserResponse, error) {
	return c.delegateVaultPlugin.NewUser(ctx, req)
}

// UpdateUser updates an existing user within the database.
func (c *CloudSQL) UpdateUser(ctx context.Context, req dbplugin.UpdateUserRequest) (dbplugin.UpdateUserResponse, error) {
	return c.delegateVaultPlugin.UpdateUser(ctx, req)
}

// DeleteUser from the database. This should not error if the user didn't
// exist prior to this call.
func (c *CloudSQL) DeleteUser(ctx context.Context, req dbplugin.DeleteUserRequest) (dbplugin.DeleteUserResponse, error) {
	return c.delegateVaultPlugin.DeleteUser(ctx, req)
}

// Type returns the Name for the particular database backend implementation.
// This type name is usually set as a constant within the database backend
// implementation, e.g. "mysql" for the MySQL database backend. This is used
// for things like metrics and logging. No behavior is switched on this.
func (c *CloudSQL) Type() (string, error) {
	return c.dbType.String(), nil
}

// Close attempts to close the underlying database connection that was
// established by the backend.
func (c *CloudSQL) Close() error {
	err := c.delegateVaultPlugin.Close()
	if err != nil {
		return err
	}
	return c.connectorCleanup()
}

func newPostgresDatabase(dbType DBType, connProducer *connutil.SQLConnectionProducer) (dbplugin.Database, func() error, func() map[string]string, error) {
	// setup the connector's "cloudsql-postgres" driver to "proxy" to cloudsql instance with Google IAM creds
	// See: https://github.com/GoogleCloudPlatform/cloud-sql-go-connector
	//
	// connection string should look like:
	// 		"host=project:region:instance user=${username} password=${password} dbname=mydb sslmode=disable"
	//
	// attribute 'sslmode=disable' is required. even though the sslmode parameter is set to disable,
	// the Cloud SQL Auth proxy does provide an encrypted connection.
	// See: https://cloud.google.com/sql/docs/postgres/connect-admin-proxy#connect-to-proxy
	cleanup, err := pgxv4.RegisterDriver(dbType.String(), cloudsqlconn.WithIAMAuthN())
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "failed to register 'postgres' driver with 'cloud-sql-go-connector'")
	}

	// delegate to vault's original postgres backend
	// See: https://github.com/hashicorp/vault/blob/main/plugins/database/postgresql/postgresql.go
	postgresBackend := &postgresql.PostgreSQL{
		SQLConnectionProducer: connProducer,
	}

	secretValues := func() map[string]string {
		return map[string]string{
			postgresBackend.Password: "[password]",
		}
	}

	return postgresBackend, cleanup, secretValues, nil
}
