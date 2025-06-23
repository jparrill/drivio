# Build stage
FROM golang:1.24.4-alpine AS builder

# Install git and ca-certificates (needed for go mod download)
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o drivio .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S drivio && \
    adduser -u 1001 -S drivio -G drivio

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/drivio .

# Change ownership to non-root user
RUN chown -R drivio:drivio /app

# Switch to non-root user
USER drivio

# Expose port (if needed)
# EXPOSE 8080

# Run the binary
ENTRYPOINT ["./drivio"]