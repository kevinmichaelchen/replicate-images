.PHONY: build test lint fmt vuln clean all

build:
	go build -o replicate-images ./cmd/replicate-images

test:
	go test -v -race -coverprofile=coverage.out ./...

lint:
	golangci-lint run

fmt:
	goimports -w .
	bunx prettier --write --prose-wrap always **/*.md

vuln:
	govulncheck ./...

clean:
	rm -f replicate-images coverage.out

all: fmt lint test build
