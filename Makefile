.DEFAULT_GOAL := test

export GOPRIVATE=github.com/kong/go-wrpc

.PHONY: install-tools
install-tools:
	./scripts/install-tools.sh

.PHONY: build
build:
	go build -o koko main.go

.PHONY: run
run:
	go run main.go

.PHONY: lint
lint:
	buf lint
	./bin/golangci-lint run ./...

.PHONY: test
test: lint
	go test -race ./...

.PHONY: gen
gen:
	./scripts/update-codegen.sh

.PHONY: gen-verify
gen-verify:
	./scripts/verify-codegen.sh

