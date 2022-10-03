.DEFAULT_GOAL := all
DEFAULT_BRANCH:=$(shell git remote show origin | sed -n '/HEAD branch/s/.*: //p')

.PHONY: install-tools
install-tools:
	./scripts/install-tools.sh

.PHONY: install-deps
install-deps:
	./scripts/install-deps.sh

.PHONY: build
build:
	go build -o koko main.go

.PHONY: run
run:
	go run main.go serve

.PHONY: lint
lint: install-deps verify-tidy
	buf format -d --exit-code
	buf lint
	./bin/golangci-lint run ./...

.PHONY: verify-tidy
verify-tidy: install-deps
	./scripts/verify-tidy.sh

.PHONY: all
all: lint test

.PHONY: test
test:
	go test -tags testsetup -count 1 ./...

test-race: install-deps
	go test -tags testsetup -count 1 -race -p 1 ./...

.PHONY: test-integration
test-integration: install-deps
	go test -tags=testsetup,integration -timeout 15m -race -count 1 -p 1 ./internal/test/...

.PHONY: gen
gen:
	./scripts/update-codegen.sh

.PHONY: gen-verify
gen-verify: install-deps
	./scripts/verify-codegen.sh

.PHONY: buf-format
buf-format:
	buf format -w

.PHONY: buf-breaking
buf-breaking:
	git fetch --no-tags origin $(DEFAULT_BRANCH)
	buf breaking --against .git#branch=origin/$(DEFAULT_BRANCH)
