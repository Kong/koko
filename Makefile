.DEFAULT_GOAL := all
DEFAULT_BRANCH:=$(shell git remote show origin | sed -n '/HEAD branch/s/.*: //p')

ATCROUTER_LIB = /tmp/lib/libatc_router.a
DEPS = $(ATCROUTER_LIB)

$(ATCROUTER_LIB):
	./scripts/build-library.sh go-atc-router/main/make-lib.sh /tmp/lib


.PHONY: install-tools
install-tools:
	./scripts/install-tools.sh

.PHONY: build
build: $(DEPS)
	go build -o koko main.go

.PHONY: run
run: $(DEPS)
	go run main.go serve

.PHONY: lint
lint: verify-tidy $(DEPS)
	buf format -d --exit-code
	buf lint
	./bin/golangci-lint run ./...

.PHONY: verify-tidy
verify-tidy:
	./scripts/verify-tidy.sh

.PHONY: all
all: lint test

.PHONY: test
test: $(DEPS)
	go test -ldflags="-extldflags=-static" -tags testsetup -count 1 ./...

test-race: $(DEPS)
	go test -ldflags="-extldflags=-static" -tags testsetup -count 1 -race -p 1 ./...

.PHONY: test-integration
test-integration: $(DEPS)
	go test -ldflags="-extldflags=-static" -tags=testsetup,integration -timeout 15m -race -count 1 -p 1 ./internal/test/...

.PHONY: gen
gen:
	./scripts/update-codegen.sh

.PHONY: gen-verify
gen-verify: $(DEPS)
	./scripts/verify-codegen.sh

.PHONY: buf-format
buf-format:
	buf format -w

.PHONY: buf-breaking
buf-breaking:
	git fetch --no-tags origin $(DEFAULT_BRANCH)
	buf breaking --against .git#branch=origin/$(DEFAULT_BRANCH)
