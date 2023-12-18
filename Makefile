.PHONY: build

DC=docker-compose

# for local development only. this plugin should be made available to your vault runtime.
# for this target to work, add a Dockerfile that builds the plugin.
image:
	docker buildx build --load --platform linux/amd64 --build-arg GIT_VERSION=local . -t gcr.io/my-registry/vault-plugin-database-cloudsql:latest

test:
	echo '{ "type": "authorized_user" }' > ./fake-creds.json
	GOOGLE_APPLICATION_CREDENTIALS=$(shell pwd)/fake-creds.json go test -race -count=1 -vet=all -v ./...

# produce a binary. for local development only. it's up to consumers to build and install the binary onto the target vault.
build:
	go get ./...
	go build ./...
	go build -o build/vault-plugin-database-cloudsql cmd/vault-plugin-database-cloudsql/*

precommit:
	pre-commit install

lint: precommit
	pre-commit run --all-files
