.PHONY: build

DC=docker-compose

# for local development only. this plugin is installed through the bootstrap sidecar.
image:
	docker build --build-arg GIT_VERSION=local . -t gcr.io/expel-engineering-devops/vault-plugin-database-cloudsql:latest

test:
	go test -race -count=1 -vet=all -v ./...

# produce a binary. for local development only. it's up to consumers to build and install the binary onto the target vault.
build:
	go get ./...
	go build ./...
	go build -o build/vault-plugin-database-cloudsql cmd/vault-plugin-database-cloudsql/*

precommit:
	pre-commit install
