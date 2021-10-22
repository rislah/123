SHELL             := /bin/bash

POSTGRES_USER     := user
POSTGRES_PASSWORD := parool
POSTGRES_HOST     := localhost
POSTGRES_PORT     := 5432
POSTGRES_DB       := user

TEST_PKGS=$(shell go list ./... | grep -v -e "integration_tests")
MIGRATE := migrate -database=postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable -path=migrations

reup: drop up

test:
	go test -race $(TEST_PKGS)

up: 
	$(MIGRATE) up

drop:
	$(MIGRATE) drop -f


integration:
	go test -race ./integration_tests
