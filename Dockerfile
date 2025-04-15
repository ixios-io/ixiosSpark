FROM golang:1.24.0-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev linux-headers git make

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN make ixiosSpark

FROM alpine:latest AS node

# Install runtime dependencies
RUN apk add --no-cache ca-certificates bash curl jq

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/build/bin/ixiosSpark /usr/local/bin/

# Create startup script for Node
RUN echo '#!/bin/bash' > /usr/local/bin/start.sh && \
    echo 'ixiosSpark --aetherBloom --syncmode full &' >> /usr/local/bin/start.sh && \
    echo 'NODE_PID=$!' >> /usr/local/bin/start.sh && \
    echo 'sleep 5' >> /usr/local/bin/start.sh && \
    echo 'if ! kill -0 $NODE_PID 2>/dev/null; then' >> /usr/local/bin/start.sh && \
    echo '    echo "Cleaning chain data due to genesis mismatch..."' >> /usr/local/bin/start.sh && \
    echo '    rm -rf /root/.ixiosSpark/aetherBloom/ixiosSpark/chaindata' >> /usr/local/bin/start.sh && \
    echo '    ixiosSpark --aetherBloom --syncmode full &' >> /usr/local/bin/start.sh && \
    echo '    NODE_PID=$!' >> /usr/local/bin/start.sh && \
    echo 'fi' >> /usr/local/bin/start.sh && \
    echo 'wait $NODE_PID' >> /usr/local/bin/start.sh && \
    chmod +x /usr/local/bin/start.sh

# Expose default ports
EXPOSE 8586 8587 38383 38383/udp

# Create data volume
VOLUME ["/root/.ixiosSpark"]

# Set the entrypoint to our startup script
ENTRYPOINT ["/usr/local/bin/start.sh"]

FROM alpine:latest AS validator

# Install runtime dependencies
RUN apk add --no-cache ca-certificates bash curl jq

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/build/bin/ixiosSpark /usr/local/bin/

# Create startup script for Validator
RUN echo '#!/bin/bash' > /usr/local/bin/start.sh && \
    echo 'if [ ! -z "$PRIVATE_KEY" ]; then' >> /usr/local/bin/start.sh && \
    echo '    # First try starting node to check for genesis error' >> /usr/local/bin/start.sh && \
    echo '    ixiosSpark --aetherBloom --gcmode=archive --maxpeers 128 --http --http.addr 127.0.0.1 --http.api personal,eth,net,web3,txpool,sealer --allow-insecure-unlock &' >> /usr/local/bin/start.sh && \
    echo '    NODE_PID=$!' >> /usr/local/bin/start.sh && \
    echo '    sleep 5' >> /usr/local/bin/start.sh && \
    echo '    if ! kill -0 $NODE_PID 2>/dev/null; then' >> /usr/local/bin/start.sh && \
    echo '        echo "Cleaning chain data due to genesis mismatch..."' >> /usr/local/bin/start.sh && \
    echo '        rm -rf /root/.ixiosSpark/aetherBloom/ixiosSpark/chaindata' >> /usr/local/bin/start.sh && \
    echo '        # Start node again after cleanup' >> /usr/local/bin/start.sh && \
    echo '        ixiosSpark --aetherBloom --http --http.addr 127.0.0.1 --http.api personal,eth,net,web3,txpool,sealer --allow-insecure-unlock &' >> /usr/local/bin/start.sh && \
    echo '    fi' >> /usr/local/bin/start.sh && \
    echo '    sleep 5' >> /usr/local/bin/start.sh && \
    echo '    # Import key if needed' >> /usr/local/bin/start.sh && \
    echo '    curl -s -H "Content-Type: application/json" -X POST --data "{\"method\": \"personal_importRawKey\", \"params\": [\"$PRIVATE_KEY\", \"\"], \"id\": 1, \"jsonrpc\": \"2.0\"}" http://localhost:8545' >> /usr/local/bin/start.sh && \
    echo '    sleep 2' >> /usr/local/bin/start.sh && \
    echo '    # Get account address' >> /usr/local/bin/start.sh && \
    echo '    ACCOUNTS=$(curl -s -H "Content-Type: application/json" -X POST --data "{\"method\": \"personal_listAccounts\", \"params\": [], \"id\": 1, \"jsonrpc\": \"2.0\"}" http://localhost:8545)' >> /usr/local/bin/start.sh && \
    echo '    ADDRESS=$(echo $ACCOUNTS | jq -r ".result[0]")' >> /usr/local/bin/start.sh && \
    echo '    # Unlock account' >> /usr/local/bin/start.sh && \
    echo '    curl -s -H "Content-Type: application/json" -X POST --data "{\"method\": \"personal_unlockAccount\", \"params\": [\"$ADDRESS\", \"\", 0], \"id\": 1, \"jsonrpc\": \"2.0\"}" http://localhost:8545' >> /usr/local/bin/start.sh && \
    echo '    # Set validator address and start sealing' >> /usr/local/bin/start.sh && \
    echo '    curl -s -H "Content-Type: application/json" -X POST --data "{\"method\": \"sealer_setEtherbase\", \"params\": [\"$ADDRESS\"], \"id\": 1, \"jsonrpc\": \"2.0\"}" http://localhost:8545' >> /usr/local/bin/start.sh && \
    echo '    curl -s -H "Content-Type: application/json" -X POST --data "{\"method\": \"sealer_start\", \"params\": [], \"id\": 1, \"jsonrpc\": \"2.0\"}" http://localhost:8545' >> /usr/local/bin/start.sh && \
    echo '    wait $NODE_PID' >> /usr/local/bin/start.sh && \
    echo 'else' >> /usr/local/bin/start.sh && \
    echo '    exec ixiosSpark --sealer.enabled --aetherBloom --gcmode archive --enable-broadcast --http --http.addr 127.0.0.1 --http.api personal,eth,net,web3,txpool,sealer' >> /usr/local/bin/start.sh && \
    echo 'fi' >> /usr/local/bin/start.sh && \
    chmod +x /usr/local/bin/start.sh

# Expose default ports
EXPOSE 8586 8587 38383 38383/udp

# Create data volume
VOLUME ["/root/.ixiosSpark"]

# Set the entrypoint to our startup script
ENTRYPOINT ["/usr/local/bin/start.sh"]
