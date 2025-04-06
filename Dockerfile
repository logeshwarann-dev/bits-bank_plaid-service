# --- Build Stage ---
FROM golang:1.23.0-alpine AS builder

WORKDIR /app

# Copy go.mod and download dependencies first for caching
COPY go.mod ./
RUN go mod download

# Copy the source code
COPY . .

# Enable static build to reduce dependencies
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# Build the binary
RUN go build -ldflags="-s -w" -o plaid-service .

# --- Final Stage ---
FROM scratch

WORKDIR /app

# Copy the static binary from the builder stage
COPY --from=builder /app/plaid-service .

# Expose ports for plaid-service connections
EXPOSE 8090

# Run the plaid-service binary
CMD ["./plaid-service"]
        