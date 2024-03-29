# Overview

This guide outlines how to quickly deploy a dev vault server,
load and configure the vault-plugin-database-cloudsql plugin,
and configure a database using terraform.

## Prerequisites

* vault-cli (version 1.8.1)
* terraform (1.2.7)

## Steps

### Build the plugin & start the server

#### Build the plugin

```bash
# From the root of this git repository
# build the plugin
make build
```

#### Create a config file for vault

```bash
# change back to the cli directory
cd quickstart/tf

# the plugin_directory is the path to the plugin binary we just built
tee vault-config.hcl <<EOF
plugin_directory = "$PWD/../../build/"
EOF
```

#### Start up a vault-server

```bash
export VAULT_ROOT_TOKEN=root

#start the server in dev mode
vault server -dev -dev-root-token-id=$VAULT_ROOT_TOKEN -log-level=debug -config=./vault-config.hcl
```

### Terraform

#### Open a new shell session

#### Export the relevant variables

```bash
export INSTANCE_NAME=test-tf-instance
export REGION=<your-region>
export GCP_PROJECT=<your-project-id>
```

#### Create a default variables file

<!-- markdownlint-disable MD013 -->
```bash
# make sure to run this from the quickstart/tf directory
tee vars.auto.tfvars <<EOF
instance_name = "$INSTANCE_NAME"
plugin_sha    = "$(sha256sum ../../build/vault-plugin-database-cloudsql | awk '{print $1}')"
region        = "$REGION"
project       = "$GCP_PROJECT"
EOF
```
<!-- markdownlint-enable MD013 -->

#### Run Terraform

```bash
terraform init
terraform apply
```

### Accessing Credentials

#### Use Vault to generate short lived database credentials

```bash
vault read cloudsql/postgres/creds/$INSTANCE_NAME
```

#### Connect to the database

```bash
# When prompted for your password paste in the value from the output of above
gcloud beta sql connect $INSTANCE_NAME \
    --user="<paste from the output above>" \
    --database=postgres
```
