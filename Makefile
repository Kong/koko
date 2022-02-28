.DEFAULT_GOAL := all

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
	go test -count 1 ./...

test-race:
	go test -v -count 1 -race -p 1 ./...

.PHONY: test-integration
test-integration:
	go test -tags=integration -race -count 1 -p 1 ./internal/test/...

.PHONY: gen
gen:
	./scripts/update-codegen.sh

.PHONY: gen-verify
gen-verify:
	./scripts/verify-codegen.sh

