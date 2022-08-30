resource "google_sql_database_instance" "test" {
  name             = var.instance_name
  database_version = "POSTGRES_12"
  region           = var.region

  // WARNING only use for dev instance
  deletion_protection = false

  settings {
    tier = "db-f1-micro"
  }
}

# generate a temporary root password for the initial connection
# will be rotated immediately after configuring the database connection
resource "random_password" "initial_connection_password" {
  length  = 16
  special = false
}

resource "google_sql_user" "vault_manager_user" {
  name     = "vault-user"
  instance = google_sql_database_instance.test.name
  # will be rotated immediately after connection configuration
  password = random_password.initial_connection_password.result
}
