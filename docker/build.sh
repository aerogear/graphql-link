#!/usr/bin/env bash
set -e
cd -P $(dirname "${BASH_SOURCE[0]}")

mkdir -p bin || true 2> /dev/null
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/graphql-link ../main.go

mkdir -p etc/graphql-link 2> /dev/null || true
cd etc/graphql-link
rm graphql-link.yaml 2> /dev/null || true
go run ../../../main.go config init
cd -

docker build -t "aerogear/graphql-link" .
docker push aerogear/graphql-link
