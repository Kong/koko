.DEFAULT_GOAL := all
DEFAULT_BRANCH:=$(shell git remote show origin | sed -n '/HEAD branch/s/.*: //p')

ATCROUTER_LIB = /tmp/lib/libatc_router.a
DEPS = $(ATCROUTER_LIB)

INSTALL_LIBS = libatc_router.so
INSTALLED_LIBS = $(addprefix /usr/lib/,$(INSTALL_LIBS))

$(ATCROUTER_LIB):
	./scripts/build-library.sh go-atc-router/main/make-lib.sh /tmp/lib

$(INSTALLED_LIBS): $(ATCROUTER_LIB)
	sudo -En ln -s /tmp/lib/$(INSTALL_LIBS) /usr/lib

.PHONY: install-tools
install-tools:
	./scripts/install-tools.sh

.PHONY: build
build: $(DEPS) $(INSTALLED_LIBS)
	go build -o koko main.go

.PHONY: run
run: $(DEPS) $(INSTALLED_LIBS)
	go run main.go serve

.PHONY: lint
lint: verify-tidy $(DEPS) $(INSTALLED_LIBS)
	buf format -d --exit-code
	buf lint
	./bin/golangci-lint run ./...

.PHONY: verify-tidy
verify-tidy:
	./scripts/verify-tidy.sh

.PHONY: all
all: lint test

.PHONY: test
test: $(DEPS) $(INSTALLED_LIBS)
	go test -tags testsetup -count 1 ./...

test-race: $(DEPS) $(INSTALLED_LIBS)
	go test -tags testsetup -count 1 -race -p 1 ./...

.PHONY: test-integration
test-integration: $(DEPS) $(INSTALLED_LIBS)
	go test -tags=testsetup,integration -timeout 15m -race -count 1 -p 1 ./internal/test/...

.PHONY: gen
gen: $(INSTALLED_LIBS)
	./scripts/update-codegen.sh

.PHONY: gen-verify
gen-verify: $(DEPS) $(INSTALLED_LIBS)
	./scripts/verify-codegen.sh

.PHONY: buf-format
buf-format:
	buf format -w

.PHONY: buf-breaking
buf-breaking:
	git fetch --no-tags origin $(DEFAULT_BRANCH)
	buf breaking --against .git#branch=origin/$(DEFAULT_BRANCH)
