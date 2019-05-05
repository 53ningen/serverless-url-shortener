.PHONY: deps clean build

build:
	GOOS=linux GOARCH=amd64 go build -o bin/create-url ./app/create-url/main
	GOOS=linux GOARCH=amd64 go build -o bin/redirect-url ./app/redirect-url/main

deps:
	dep ensure

clean: 
	rm -rf ./bin/*
