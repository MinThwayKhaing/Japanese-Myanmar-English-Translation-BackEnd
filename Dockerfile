# Start from official Go image
FROM golang:1.25-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum first (for caching dependencies)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the binary
RUN go build -o main ./cmd/main.go

# ---- Runtime Stage ----
FROM alpine:latest

WORKDIR /app

# Copy built binary from builder
COPY --from=builder /app/main .

# Copy static frontend if exists
COPY web ./web

# Expose port
EXPOSE 8085

# Run the binary
CMD ["./main"]
