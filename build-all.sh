#!/bin/bash

GOARCH=arm64 GOOS=linux go build -ldflags="-s -w" -o "./build/bin/ixiosSpark-arm64" ./cmd/ixiosSpark
GOARCH=amd64 GOOS=linux go build -ldflags="-s -w" -o "./build/bin/ixiosSpark-amd64" ./cmd/ixiosSpark
GOARCH=amd64 GOOS=windows go build -ldflags="-s -w" -o "./build/bin/ixiosSpark-amd64.exe" ./cmd/ixiosSpark
GOARCH=arm64 GOOS=windows go build -ldflags="-s -w" -o "./build/bin/ixiosSpark-arm64.exe" ./cmd/ixiosSpark