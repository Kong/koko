#!/bin/bash -e

buf generate \
  --template ./internal/grpc/proto/buf.gen.yaml \
  --path internal/grpc/proto/kong \
  --timeout 5m
buf generate \
  --template ./internal/wrpc/proto/buf.gen.yaml \
  --path internal/wrpc/proto/kong \
  --timeout 5m
go generate -ldflags="-extldflags=-static" ./...
