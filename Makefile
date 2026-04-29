.PHONY: test build lint install

test:
	go test ./...

build:
	go build ./cmd/structtagger

lint:
	golangci-lint run ./...

install:
	go install ./cmd/structtagger
