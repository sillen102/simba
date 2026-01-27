# Multi-stage Dockerfile for Simba telemetry example

# Stage 1: Build the application
FROM golang:1.25-alpine AS builder

WORKDIR /build

# Copy the entire simba library (since we use replace directive)
COPY ../../ /simba

# Set working directory to the example
WORKDIR /simba/examples/telemetry

# Download dependencies
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o telemetry-demo .

# Stage 2: Create minimal runtime image
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /simba/examples/telemetry/telemetry-demo .

# Expose the application port
EXPOSE 9999

# Run the application
CMD ["./telemetry-demo"]
