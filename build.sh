#!/bin/sh
# This script is to prepare the release. To just run a modified version,
# you can just run `go run mapshot.go` (see README.md).
go generate ./...
go test ./... || exit 1

TARGET="$(pwd)/build"
mkdir -p "${TARGET?}"
GOOS=linux GOARCH=amd64 go build -o "${TARGET?}/mapshot-linux-amd64" mapshot.go
GOOS=windows GOARCH=amd64 go build -o "${TARGET?}/mapshot-windows-amd64.exe" mapshot.go
GOOS=darwin GOARCH=amd64 go build -o "${TARGET?}/mapshot-darwin-amd64" mapshot.go
go run mapshot.go package "${TARGET?}"