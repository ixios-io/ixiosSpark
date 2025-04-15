.PHONY: ixiosSpark all test clean devtools help

GOBIN = ./build/bin

# Build the ixiosSpark binary
ixiosSpark:
	@mkdir -p $(GOBIN)
	GO111MODULE=on go build -o $(GOBIN)/ixiosSpark ./cmd/ixiosSpark
	@echo "Done building ixiosSpark."
	@echo "Run \"$(GOBIN)/ixiosSpark\" to launch ixiosSpark."

# Build all main packages
all:
	@mkdir -p $(GOBIN)
	GO111MODULE=on go build -o $(GOBIN)/ixiosSpark ./cmd/ixiosSpark
	GO111MODULE=on go build -o $(GOBIN)/evm     ./cmd/evm
	GO111MODULE=on go build -o $(GOBIN)/rlpdump ./cmd/rlpdump
	@echo "Done building all main packages."

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Clean up build artifacts and caches
clean:
	@echo "Cleaning..."
	go clean -cache
	rm -rf $(GOBIN)

# Install recommended developer tools
devtools:
	env GOBIN= go install golang.org/x/tools/cmd/stringer@latest
	env GOBIN= go install github.com/fjl/gencodec@latest
	env GOBIN= go install github.com/golang/protobuf/protoc-gen-go@latest
	env GOBIN= go install ./cmd/abigen
	@type "solc" 2> /dev/null || echo "Please install solc"
	@type "protoc" 2> /dev/null || echo "Please install protoc"

# Show help information
help:
	@echo "Available Make targets:"
	@echo "  ixiosSpark    Build the ixiosSpark binary"
	@echo "  all        Build all main packages"
	@echo "  test       Run tests"
	@echo "  clean      Remove built binaries and clean Go caches"
	@echo "  devtools   Install recommended developer tools"
	@echo "  help       Show this help message"