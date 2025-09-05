# Build stage
FROM golang:1.25-alpine AS builder

# Install git for Go modules that require it
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o pingo \
    ./main.go

#####################
# Build final image #
#####################
FROM alpine:latest

# Install ca-certificates for HTTPS requests and tzdata for timezone
RUN apk --no-cache add ca-certificates tzdata && \
    update-ca-certificates

# Create a non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/pingo .

# Copy any additional files needed (like migrations if required at runtime)
# Uncomment the next line if migrations are needed in the container
# COPY --from=builder /app/migrations ./migrations

# Change ownership of the app directory to the non-root user
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port (adjust as needed based on your application)
EXPOSE 8080

# Specify the container's entrypoint
ENTRYPOINT ["./pingo"]