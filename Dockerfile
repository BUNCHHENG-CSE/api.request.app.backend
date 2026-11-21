# Stage 1: Build the Go binary
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copy dependency files and download them
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build statically compiled binary with stripped symbols for minimal size and fast boot
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /api-server ./cmd/api/main.go

# Stage 2: Create a minimal production image
FROM alpine:latest

# Install CA certificates for outbound HTTPS request execution and tzdata for timezones
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Create a non-root system user and prepare local data directories
RUN addgroup -S appgroup && adduser -S appuser -G appgroup \
    && mkdir -p /data/collections \
    && chown -R appuser:appgroup /data/collections /app

# Copy the binary from the builder stage
COPY --from=builder /api-server /app/api-server

USER appuser

EXPOSE 8080

CMD ["/app/api-server"]
