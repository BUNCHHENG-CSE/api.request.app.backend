# Stage 1: Build the Go binary
FROM golang:1.26-alpine AS builder

# Set the working directory
WORKDIR /app

# Copy dependency files and download them
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application (disabling CGO for static compilation)
RUN CGO_ENABLED=0 GOOS=linux go build -o /api-server ./cmd/api/main.go

# Stage 2: Create a minimal production image
FROM alpine:latest

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /api-server .

# Ensure the app can read the .env file if it exists locally
COPY .env . 

# Expose the application port
EXPOSE 8080

# Run the binary
CMD ["./api-server"]