#!/bin/bash

buf generate --template ./internal/grpc/proto/buf.gen.yaml --path internal/grpc/proto/kong
buf generate --template ./internal/wrpc/proto/buf.gen.yaml --path internal/wrpc/proto/kong
