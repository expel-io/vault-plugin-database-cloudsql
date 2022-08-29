# register the cloudsql plugin
resource "vault_generic_endpoint" "configure_custom_cloudsql_plugin" {
  path                 = "sys/plugins/catalog/database/vault-plugin-database-cloudsql"
  disable_read         = false
  disable_delete       = true
  ignore_absent_fields = true

  data_json = jsonencode({
    type    = "database"
    sha_256 = var.plugin_sha
    command = "vault-plugin-database-cloudsql"
    // plugin runs as an rpc service on an unix socket.
    // set TMPDIR to ensure it can be opened on a filesystem that vault can write to.
    // See: https://github.com/hashicorp/go-plugin/blob/5dee41c45dcec4c3b00c6d227bec8e5ebea56a8a/server.go#L546
    env = ["TMPDIR=/tmp"]

    // the follwoing args will be passed to the plugin binary
    // See: https://www.vaultproject.io/api-docs/system/plugins-catalog#args
    args = ["-db-type=cloudsql-postgres", "-log-level=${local.cloudsql_plugin_log_level}"]
  })
}

# mount a database path for all instances of type cloudsql/postgres
resource "vault_mount" "cloudsql_postgres_mount" {
  path        = "cloudsql/postgres"
  type        = "database"
  description = "Google CloudSQL postgres database keys"

  # see: https://www.vaultproject.io/api/system/mounts#options
  options = {
    plugin_name = "vault-plugin-database-cloudsql"
  }

  default_lease_ttl_seconds = local.access_default_ttl_seconds
  max_lease_ttl_seconds     = local.access_max_ttl_seconds
}
