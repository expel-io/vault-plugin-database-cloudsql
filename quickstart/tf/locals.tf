locals {
  database_name = "postgres"
  # attribute 'sslmode=disable' is required. even though the sslmode parameter is set to disable,
  # the Cloud SQL Auth proxy does provide an encrypted connection.
  # See: https://cloud.google.com/sql/docs/postgres/connect-admin-proxy#connect-to-proxy
  sql_connection_format = "host=%s:%s:%s dbname=%s port=%s user={{username}} password={{password}} sslmode=disable"
  sql_connection_url = format(
    local.sql_connection_format,
    var.project,
    var.region,
    var.instance_name,
    local.database_name,
    var.instance_port
  )

  # default 1hr
  access_default_ttl_seconds = 3600
  # allow extending up to 6 hours
  access_max_ttl_seconds    = 21600
  cloudsql_plugin_log_level = "debug"
}
