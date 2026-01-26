.PHONY: build test test-e2e lint fmt vuln clean all

build:
	go build -o replicate-images ./cmd/replicate-images

test:
	go test -v -race -coverprofile=coverage.out ./...

test-e2e:
	@echo "WARNING: This may cost money (1 image if not cached)"
	@./scripts/test-e2e.sh

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
