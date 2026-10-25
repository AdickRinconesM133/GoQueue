.PHONY: run build test 

run:
	go run ./cmd/api

build:
	go build -o bin/app ./cmd/api
