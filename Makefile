.PHONY: deps clean build

build:
	GOOS=linux GOARCH=amd64 go build -o bin/create-url ./app/create-url
	GOOS=linux GOARCH=amd64 go build -o bin/redirect-url ./app/redirect-url

deps:
	dep ensure

clean: 
	rm -rf ./bin/*
