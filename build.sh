#!/bin/sh
# This script is to prepare the release. To just run a modified version,
# you can just run `go run mapshot.go` (see README.md).
./generate.sh
go test ./... || exit 1

TARGET="$(pwd)/build"
mkdir -p "${TARGET?}"
GOOS=linux GOARCH=amd64 go build -o "${TARGET?}/mapshot-linux" mapshot.go
GOOS=linux GOARCH=arm go build -o "${TARGET?}/mapshot-linux-arm" mapshot.go
GOOS=windows GOARCH=amd64 go build -o "${TARGET?}/mapshot-windows.exe" mapshot.go
GOOS=darwin GOARCH=amd64 go build -o "${TARGET?}/mapshot-darwin" mapshot.go
GOOS=darwin GOARCH=arm64 go build -o "${TARGET?}/mapshot-darwin-arm64" mapshot.go
go run mapshot.go package "${TARGET?}"
