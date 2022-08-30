# Overview

This guide outlines how to quickly deploy a dev vault server,
load and configure the vault-plugin-database-cloudsql plugin,
and configure a database.

## Prerequisites

* vault-cli (version 1.8.1)
* gcloud

## Steps

### Create and configure a database in GCP

#### Ensure you have your [application default credentials][0] set

#### Create the instance

```bash
export INSTANCE_NAME=<instance-name>
export REGION=<region>
export DB_VERSION=<pg-version> #for example, POSTGRES_14
export GCP_PROJECT=<your-project-id>

gcloud sql instances create $INSTANCE_NAME \
--tier=db-f1-micro \
--region=$REGION \
--database-version=$DB_VERSION
```

#### Create an admin user in the instance

```bash

export VAULT_DB_USER=vault-user
export VAULT_DB_USER_PASS=ilovemymom123

# vault will use this user to connect to the instance
gcloud sql users create $VAULT_DB_USER \
-i $INSTANCE_NAME \
--password=$VAULT_DB_USER_PASS
```

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
cd quickstart/cli

# the plugin_directory is the path to the plugin binary we just built
tee vault-config.hcl <<EOF
plugin_directory = "$PWD/../../build/"
EOF
```

#### Start up a vault-server

```bash
export VAULT_ROOT_TOKEN=root

vault server -dev -dev-root-token-id=$VAULT_ROOT_TOKEN -log-level=debug -config=./vault-config.hcl
```

### Configure the plugin

#### Open a new shell session

#### Login to the vault server

```bash
export VAULT_ROOT_TOKEN=root
export VAULT_ADDR='http://127.0.0.1:8200'

vault login $VAULT_ROOT_TOKEN
```

#### List the plugins

```bash
# Notice there is no vault-plugin-database-cloudsql plugin
vault plugin list
```

#### Register the plugin

<!-- markdownlint-disable MD013 -->
```bash
tee register-payload.json <<EOF
{
    "name": "vault-plugin-database-cloudsql",
    "type": "database",
    "sha256": "$(sha256sum vault-plugin-database-cloudsql/build/vault-plugin-database-cloudsql | awk '{print $1}')",
    "args": [
        "-log-level=debug",
        "-db-type=cloudsql-postgres"
    ],
    "command": "vault-plugin-database-cloudsql",
    "env": [
        "TMPDIR=/tmp"
    ]
}
EOF

curl --header "X-Vault-Token: $VAULT_ROOT_TOKEN" \
    --request POST \
    --data @register-payload.json \
    http://127.0.0.1:8200/v1/sys/plugins/catalog/database/vault-plugin-database-cloudsql
```
<!-- markdownlint-enable MD013 -->

#### List the plugins again

```bash
# Notice there is no vault-plugin-database-cloudsql plugin
vault plugin list | grep vault-plugin-database-cloudsql
```

#### Get some info about the plugin

```bash
vault plugin info database vault-plugin-database-cloudsql
```

#### List the secrets

```bash
# Notice there is no mount for the plugin we just configured
vault secrets list
```

#### Mount and enable the secrets backend

```bash
tee mount-payload.json <<EOF
{
    "path": "/cloudsql/postgres",
    "type": "database",
    "description": "",
    "config": {
        "default_lease_ttl": "5m",
        "max_lease_ttl": "30m"
    },
    "options": {
        "plugin_name": "vault-plugin-database-cloudsql"
    }
}
EOF

curl --header "X-Vault-Token: $VAULT_ROOT_TOKEN" \
    --request POST \
    --data @mount-payload.json \
    http://127.0.0.1:8200/v1/sys/mounts/cloudsql/postgres
```

#### List the secrets backends again

```bash
# Notice the /cloudsql/postgres mount
vault secrets list
```

### Configure Vault for the specific instance

#### Export all of the necessary variables

```bash
export INSTANCE_NAME=test-instance
export REGION=us-east1
export DB_VERSION=POSTGRES_12
export GCP_PROJECT=<your-project-id>
export VAULT_DB_USER=vault-user
export VAULT_DB_USER_PASS=ilovemymom123
```

#### Create a vault-managed connection for the instance

<!-- markdownlint-disable MD013 -->
```bash
vault write cloudsql/postgres/config/test-instance \
    plugin_name="vault-plugin-database-cloudsql" \
    allowed_roles="test-instance" \
    connection_url="host=$GCP_PROJECT:$REGION:$INSTANCE_NAME dbname=postgres port=5432 user={{username}} password={{password}} sslmode=disable" \
    username="$VAULT_DB_USER" \
    password="$VAULT_DB_USER_PASS"
```
<!-- markdownlint-enable MD013 -->

#### Create a default readonly database accessor role

<!-- markdownlint-disable MD013 -->
```bash
vault write cloudsql/postgres/roles/test-instance \
    db_name=test-instance \
    creation_statements="CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}'; \
        GRANT SELECT ON ALL TABLES IN SCHEMA public TO \"{{name}}\";" \
    default_ttl="1h" \
    max_ttl="24h"
```
<!-- markdownlint-enable MD013 -->

### Accessing Credentials

#### Use Vault to generate short lived database credentials

```bash
vault read cloudsql/postgres/creds/test-instance
```

#### Connect to the database

```bash
# When prompted for your password paste in the value from the output of above
gcloud beta sql connect $INSTANCE_NAME \
    --user="<paste from the output above>" \
    --database=postgres
```

[0]: https://cloud.google.com/sdk/gcloud/reference/auth/application-default
