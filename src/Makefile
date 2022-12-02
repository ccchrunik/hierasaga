
.PHONY: build run test install

install:
	go get github.com/rakyll/gotest
	go install

build:
	go build .

run:
	go run .

test:
	gotest -v ./...