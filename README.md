# vault-plugin-database-cloudsql

This is a Hashicorp Vault database plugin to connect to CloudSQL instances with [GoogleCloudPlatform/cloud-sql-go-connector][0].

From Google Cloud's documentation:
<!-- markdownlint-disable MD013 -->
> Using the Cloud SQL Auth proxy is the recommended method for connecting to a Cloud SQL instance. See: [Connect using the Cloud SQL Auth proxy][0]
<!-- markdownlint-enable MD013 -->

By using the [cloud-sql-go-connector][0] Hashicorp Vault is able to connect to
multiple CloudSQL instances without the need for the [Cloud SQL Auth Proxy][2].

This plugin does two things:

1. Initializes the database driver with the [cloud-sql-go-connector][0]
allowing it to connect securely with Google IAM credentials.
2. It then defers to Hashicorp Vault's [original database plugins][3]
for all database specific interactions.

**NOTE:** Currently support is limited to Postgres instances.

## Arguments

The following plugin arguments are supported:

* `-db-type`, defaults to `cloudsql-postgres`.
This is currently the only supported database type.
* `-log-level`, defaults to `info`

## Example Usage with Terraform

After building this plugin and making it available to your Vault
runtime, you can [register][4] the plugin to the [plugin catalog][5] like this:

```hcl
resource "vault_generic_endpoint" "configure_custom_cloudsql_plugin" {
  path                 = "sys/plugins/catalog/database/vault-plugin-database-cloudsql"
  disable_read         = false
  disable_delete       = false
  ignore_absent_fields = true

  data_json = jsonencode({
    type    = "database"
    sha_256 = <INSERT-YOUR-BINARY-SHA>
    command = "vault-plugin-database-cloudsql"
    args = [
        "-db-type=cloudsql-postgres",
        "-log-level=info"
    ]
  })
}
```

[0]: github.com/GoogleCloudPlatform/cloud-sql-go-connector
[1]: https://cloud.google.com/sql/docs/postgres/connect-admin-proxy#overview
[2]: https://cloud.google.com/sql/docs/postgres/sql-proxy
[3]: https://github.com/hashicorp/vault/tree/main/plugins/database
[4]: https://www.vaultproject.io/docs/plugins/plugin-architecture#plugin-registration
[5]: https://www.vaultproject.io/api-docs/system/plugins-catalog
