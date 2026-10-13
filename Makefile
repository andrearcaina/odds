.PHONY: build run test fmt fmt-check vet tidy check clean

build:
	go build -o bin/odds ./cmd/server

run: build
	./bin/odds

test:
	go test ./...

fmt:
	gofmt -w cmd internal

fmt-check:
	test -z "$$(gofmt -l cmd internal)"

vet:
	go vet ./...

tidy:
	go mod tidy

check: fmt-check tidy test vet build

clean:
	rm -rf bin
