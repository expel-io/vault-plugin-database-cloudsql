# create a vault-managed connection for the instance.
# resource 'vault_database_secret_backend_connection' does not play nicely
# when using a custom plugin. use a generic endpoint instead.
resource "vault_generic_endpoint" "configure_tenant_database_connection" {
  path                 = "${var.cloudsql_postgres_mount_path}/config/${google_sql_database_instance.test.name}"
  disable_read         = false
  disable_delete       = true
  ignore_absent_fields = true

  data_json = jsonencode({
    plugin_name       = "vault-plugin-database-cloudsql"
    connection_url    = local.sql_connection_url
    verify_connection = true
    allowed_roles     = google_sql_database_instance.test.name
    username          = google_sql_user.vault_manager_user.name
    # this is a throw-away password. it's about to be auto-rotated by vault and kept secret
    password = google_sql_user.vault_manager_user.password
  })
}

# rotate the random root password
# See: https://www.vaultproject.io/api-docs/secret/databases#rotate-root-credentials
resource "vault_generic_endpoint" "rotate_initial_db_password" {
  depends_on     = [vault_generic_endpoint.configure_tenant_database_connection]
  path           = "${var.cloudsql_postgres_mount_path}/rotate-root/${google_sql_database_instance.test.name}"
  disable_read   = true
  disable_delete = true
  data_json      = "{}"
}

# Install a default readonly database accessor
resource "vault_database_secret_backend_role" "default_readonly_role_database_accessor" {
  name        = google_sql_database_instance.test.name
  backend     = var.cloudsql_postgres_mount_path
  db_name     = google_sql_database_instance.test.name
  default_ttl = local.access_default_ttl_seconds
  max_ttl     = local.access_max_ttl_seconds
  creation_statements = [
    // creates role
    "CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}';",
    // grants readonly permission
    "GRANT SELECT ON ALL TABLES IN SCHEMA public TO \"{{name}}\" ;"
  ]
  revocation_statements = [
    // revokes readonly permissions
    "REVOKE SELECT ON ALL TABLES IN SCHEMA public FROM \"{{name}}\" ;",
    // DROP the ROLE
    "DROP ROLE \"{{name}}\";",
  ]
}
