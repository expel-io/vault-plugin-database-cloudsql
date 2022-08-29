terraform {
  required_providers {
    vault = {
      source = "hashicorp/vault"
    }
    google = {
      source = "hashicorp/google"
    }
  }
}

provider "vault" {
  address = "http://127.0.0.1:8200"
}

provider "google" {
  project = var.project
  region  = var.region
}
