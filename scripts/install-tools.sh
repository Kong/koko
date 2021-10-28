#!/bin/bash -e

# check go installed
go version > /dev/null 2>&1
if [[ $? -ne 0 ]];
then
  echo "Please install go toolchain on the system."
  exit 1
fi

# check docker installed
docker version > /dev/null 2>&1
if [[ $? -ne 0 ]];
then
  echo "Please install docker on the system."
  :wqa
  exit 1
fi

go install "github.com/bufbuild/buf/cmd/buf"
go install "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway"
go install "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2"
go install "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
go install "github.com/kong/go-wrpc/cmd/protoc-gen-go-wrpc"
go install "google.golang.org/protobuf/cmd/protoc-gen-go"

# install golangci-lint if needed

GOLANGCI_LINT_VERSION=1.42.1
FOUND_LINT_VERSION=$(./bin/golangci-lint version --format short)
if [[ "$FOUND_LINT_VERSION" != "${GOLANGCI_LINT_VERSION}" ]];
then
  curl -sSfL \
  "https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh" \
  | sh -s -- -b bin v${GOLANGCI_LINT_VERSION}
fi
