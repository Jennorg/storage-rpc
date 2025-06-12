#!/bin/bash

set -e

PROTO_SRC_DIR="pkg/api"
PROTOC_GEN_GO="/home/jenorg/go/bin/protoc-gen-go"
PROTOC_GEN_GO_GRPC="/home/jenorg/go/bin/protoc-gen-go-grpc"

echo "Generating gRPC and Protobuf code from ${PROTO_SRC_DIR}/*.proto..."

protoc \
  --proto_path=${PROTO_SRC_DIR} \
  --plugin=protoc-gen-go=${PROTOC_GEN_GO} \
  --plugin=protoc-gen-go-grpc=${PROTOC_GEN_GO_GRPC} \
  --go_out=${PROTO_SRC_DIR} --go_opt=paths=source_relative \
  --go-grpc_out=${PROTO_SRC_DIR} --go-grpc_opt=paths=source_relative \
  ${PROTO_SRC_DIR}/*.proto

echo "Protobuf generation complete."
