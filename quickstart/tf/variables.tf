variable "instance_name" {
  description = "name of the cloudsql instance to create"
}
variable "plugin_sha" {
  description = "hash of the plugin binary, to generate run: $ sha256sum /path/to/plugin/binary' | awk '{print $1}'"
}

variable "project" {
    description = "Google Cloud project name"
  
}

variable "cloudsql_postgres_mount_path" {
  default = "cloudsql/postgres"
}

variable "region" {
    default = "us-east1"
}
  
variable "instance_port" {
    default = "5432"
}



