# Build stage
FROM golang:alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o metrics-service ./cmd/api

# Final stage
FROM alpine:latest

WORKDIR /app

# Install certificates
RUN apk --no-cache add ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /app/metrics-service .
COPY --from=builder /app/docs ./docs

# Expose the port
EXPOSE 8085

# Run the binary
CMD ["./metrics-service"]
