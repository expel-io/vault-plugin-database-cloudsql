# Overview

This guide outlines how to quickly deploy a dev vault server,
load and configure the vault-plugin-database-cloudsql plugin,
and configure a database using terraform.

## Prerequisites

* vault-cli (version 1.8.1)
* terraform

## Steps

### Build the plugin

* Build the plugin

```bash
# change to the git root directory
cd ../../

# build the plugin
make build
```

* Create a config file for vault.

```bash
# change back to the cli directory
cd quickstart/cli

# the plugin_directory is the path to the plugin binary we just built
tee vault-config.hcl <<EOF
plugin_directory = "$PWD/../../build/"
EOF
```

* Start up a vault-server

```bash
#export the relevant variables
export VAULT_ROOT_TOKEN=root

#start the server in dev mode
vault server -dev -dev-root-token-id=$VAULT_ROOT_TOKEN -log-level=debug -config=./vault-config.hcl
```

## Terraform

### IN A NEW TERMINAL

* Export the relevant variables

```bash
export INSTANCE_NAME=test-tf-instance
export REGION=<your-region>
export GCP_PROJECT=<your-project-id>
```

* Create a default variables file

<!-- markdownlint-disable MD013 -->
```bash
tee vars.auto.tfvars <<EOF
instance_name = "$INSTANCE_NAME"
plugin_sha    = "$(sha256sum vault-plugin-database-cloudsql/build/vault-plugin-database-cloudsql | awk '{print $1}')"
region        = "$REGION"
project       = "$GCP_PROJECT"
EOF
```
<!-- markdownlint-enable MD013 -->

* Run Terraform

```bash
terraform init
terraform apply
```

## Accessing Credentials

* Use Vault to generate short lived database credentials!

```bash
vault read cloudsql/postgres/creds/$INSTANCE_NAME
```

* Connect to the database.

```bash
# When prompted for your password paste in the value from the output of above
gcloud beta sql connect $INSTANCE_NAME \
    --user="<paste from the output above>" \
    --database=postgres --port=5433
```
