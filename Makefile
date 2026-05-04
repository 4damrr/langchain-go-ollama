#include .env

swagger:
	swag init --pd -g main.go -o docs

#migrate-create:
#	goose create $(action) sql
#
#migrate-up:
#	goose up

#migrate-down:
#	goose down

#generate-mock:
#	mockgen -source=internal/$(path)/repository.go -destination=internal/$(path)/repository_mock.go -package=$(shell basename $(path))
#	mockgen -source=internal/$(path)/usecase.go -destination=internal/$(path)/usecase_mock.go -package=$(shell basename $(path))

run:
	go run main.go
