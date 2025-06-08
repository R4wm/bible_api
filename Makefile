# Makefile for bible_api project

.PHONY: all build run test clean

all: build

build:
	go build -o bible_api ./cmd/bible_api.go

run: build
	./bible_api

test:
	go test ./...

clean:
	rm -f bible_api
