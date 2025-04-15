#!/bin/bash

mkdir -p ./build

find_files() {
  find . ! \( \
      \( \
        -path '.github' \
        -o -path './build/_workspace' \
        -o -path './build/bin' \
        -o -path './crypto/bn256' \
        -o -path '*/vendor/*' \
      \) -prune \
    \) -name '*.go'
}

# Try to find goimports in various locations
if command -v goimports &> /dev/null; then
    GOIMPORTS="goimports -w"
elif [ -f "$HOME/go/bin/goimports" ]; then
    GOIMPORTS="$HOME/go/bin/goimports -w"
else
    echo "Error: goimports not found. install with 'go install golang.org/x/tools/cmd/goimports@latest'"
    exit 1
fi

GOFMT="gofmt -s -w"
find_files | xargs $GOIMPORTS
find_files | xargs $GOFMT

go mod tidy

make clean
go clean -cache
make ixiosSpark