.DEFAULT_GOAL := all

export GOPRIVATE=github.com/kong/go-wrpc

.PHONY: install-tools
install-tools:
	./scripts/install-tools.sh

.PHONY: build
build:
	go build -o koko main.go

.PHONY: run
run:
	go run main.go serve

.PHONY: lint
lint:
	buf lint
	./bin/golangci-lint run ./...

.PHONY: all
all: lint test

.PHONY: test
test:
	go test -race ./...

.PHONY: test-integration
test-integration:
	go test -tags=integration ./internal/test/...

.PHONY: gen
gen:
	./scripts/update-codegen.sh

.PHONY: gen-verify
gen-verify:
	./scripts/verify-codegen.sh

