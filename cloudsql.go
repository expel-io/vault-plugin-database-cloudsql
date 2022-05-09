package cloudsql

import (
	"context"

	"cloud.google.com/go/cloudsqlconn"
	"cloud.google.com/go/cloudsqlconn/postgres/pgxv4"
	"github.com/hashicorp/vault/plugins/database/postgresql"
	dbplugin "github.com/hashicorp/vault/sdk/database/dbplugin/v5"
	"github.com/hashicorp/vault/sdk/database/helper/connutil"
)

var (
	DbType = "cloudsql-postgres"
)

// See: https://www.vaultproject.io/docs/secrets/databases/custom
type CloudSQL struct {
	*connutil.SQLConnectionProducer

	// TODO rename to communicate that it is a delegate object
	p *postgresql.PostgreSQL

	cloudsqlConnectorCleanup func() error
}

func New() (interface{}, error) {
	// setup the connector's "cloudsql-postgres" driver to "proxy" to cloudsql instance with Google IAM creds
	// See: https://github.com/GoogleCloudPlatform/cloud-sql-go-connector
	//
	// connection string should look like:
	// 	"host=project:region:instance user=${username} password=${password} dbname=mydb ssldisable=true"
	// TODO document why ssldisable=true
	cleanup, err := pgxv4.RegisterDriver(DbType, cloudsqlconn.WithIAMAuthN())
	if err != nil {
		return nil, err
	}

	connProducer := &connutil.SQLConnectionProducer{}
	connProducer.Type = DbType

	// delegate to vault's original postgres backend
	// See: https://github.com/hashicorp/vault/blob/main/plugins/database/postgresql/postgresql.go
	p := &postgresql.PostgreSQL{
		SQLConnectionProducer: connProducer,
	}

	db := &CloudSQL{
		SQLConnectionProducer:    connProducer,
		p:                        p,
		cloudsqlConnectorCleanup: cleanup,
	}
	// Wrap the plugin with middleware to sanitize errors
	wrappedDb := dbplugin.NewDatabaseErrorSanitizerMiddleware(db, db.secretValues)
	return wrappedDb, nil
}

// Initialize the database plugin. This is the equivalent of a constructor for the
// database object itself.
func (c *CloudSQL) Initialize(ctx context.Context, req dbplugin.InitializeRequest) (dbplugin.InitializeResponse, error) {
	return c.p.Initialize(ctx, req)
}

// NewUser creates a new user within the database. This user is temporary in that it
// will exist until the TTL expires.
func (c *CloudSQL) NewUser(ctx context.Context, req dbplugin.NewUserRequest) (dbplugin.NewUserResponse, error) {
	return c.p.NewUser(ctx, req)
}

// UpdateUser updates an existing user within the database.
func (c *CloudSQL) UpdateUser(ctx context.Context, req dbplugin.UpdateUserRequest) (dbplugin.UpdateUserResponse, error) {
	return c.p.UpdateUser(ctx, req)
}

// DeleteUser from the database. This should not error if the user didn't
// exist prior to this call.
func (c *CloudSQL) DeleteUser(ctx context.Context, req dbplugin.DeleteUserRequest) (dbplugin.DeleteUserResponse, error) {
	return c.p.DeleteUser(ctx, req)
}

// Type returns the Name for the particular database backend implementation.
// This type name is usually set as a constant within the database backend
// implementation, e.g. "mysql" for the MySQL database backend. This is used
// for things like metrics and logging. No behavior is switched on this.
func (c *CloudSQL) Type() (string, error) {
	return DbType, nil
}

// Close attempts to close the underlying database connection that was
// established by the backend.
func (c *CloudSQL) Close() error {
	err := c.p.Close()
	if err != nil {
		return err
	}
	return c.cloudsqlConnectorCleanup()
}

func (c *CloudSQL) secretValues() map[string]string {
	return map[string]string{
		c.p.Password: "[password]",
	}
}
