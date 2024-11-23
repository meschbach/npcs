#!/bin/bash

go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.2
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1

function gen_grpc() {
  file="$1"
  protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    "$file"
}

gen_grpc t3/network/game.proto
gen_grpc competition/wire/competition.proto
