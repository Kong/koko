#!/bin/bash

# check go installed
go version > /dev/null 2>&1
if [[ $? -ne 0 ]];
then
  echo "Please install go toolchain on the system."
  exit 1
fi


go install "github.com/bufbuild/buf/cmd/buf"
go install "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway"
go install "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2"
go install "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
go install "github.com/kong/go-wrpc/cmd/protoc-gen-go-wrpc"
go install "google.golang.org/protobuf/cmd/protoc-gen-go"

curl -sSfL \
"https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh" \
| sh -s -- -b bin v1.41.1
