.PHONY: build test lint clean

build:
	go build -o outrunner ./cmd/outrunner

test:
	go test ./...

lint:
	golangci-lint run

cross:
	GOOS=linux GOARCH=amd64 go build -o outrunner-linux-amd64 ./cmd/outrunner
	GOOS=linux GOARCH=arm64 go build -o outrunner-linux-arm64 ./cmd/outrunner
	GOOS=darwin GOARCH=arm64 go build -o outrunner-darwin-arm64 ./cmd/outrunner

clean:
	rm -f outrunner outrunner-linux-* outrunner-darwin-*
