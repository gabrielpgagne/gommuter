# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w -s' -o gommutetime .

# Runtime stage
FROM alpine:3.19

# Install ca-certificates for HTTPS and tzdata for timezone support
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/gommutetime /app/gommutetime

# Create data directory
RUN mkdir -p /app/data

# Config file will be mounted as volume
VOLUME ["/app/data"]

# Run scheduler by default
CMD ["/app/gommutetime", "schedule", "-config", "/app/config.yaml"]
