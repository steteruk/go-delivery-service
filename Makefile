.PHONY: install create-migration migrate run run-courier run-order run-location run-location-track

MIGRATION_NAME ?= $(shell bash -c 'read  -p "Enter Migration Name: " migrationName; echo $$migrationName')
SERVICE_NAME ?= $(shell bash -c 'read -p "Enter Service Name: " serviceName; echo $$serviceName')
DB_NAME ?= $(shell bash -c 'read  -p "Enter Db Name: " dbName; echo $$dbName')


install:
	docker-compose build
	go install github.com/pressly/goose/v3/cmd/goose@latest
	go install github.com/gojuno/minimock/v3/cmd/minimock@latest

create-migration:
	goose -dir "./$(SERVICE_NAME)/db/migrations"  create $(MIGRATION_NAME) sql

migrate:
	goose -dir "./$(SERVICE_NAME)/db/migrations" postgres "host=localhost password=S3cret  user=citizix_user dbname=${DB_NAME} sslmode=disable" up-by-one

down-migrate:
	goose -dir "./$(SERVICE_NAME)/db/migrations" postgres "host=localhost password=S3cret  user=citizix_user dbname=${DB_NAME} sslmode=disable" down

run:
	make -j 4 run-courier run-order run-location run-location-track

run-courier:
	cd courier/cmd/courier/ && go run main.go

run-order:
	cd order/cmd/order/ && go run main.go

run-location:
	cd location/cmd/location/ && go run main.go

run-location-track:
	cd location/cmd/track/ && go run main.go
