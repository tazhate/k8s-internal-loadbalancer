# First Stage: Build the Go binary
FROM golang:1.24-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY *.go ./
COPY pkg/ ./pkg/

# Build arguments for versioning
ARG VERSION=dev
ARG BUILD_TIME
ARG VCS_REF

# Build the Go application with optimizations and version info
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.VCSRef=${VCS_REF}" \
    -o traefik-updater

# Second Stage: Create a minimal image with the Go binary
FROM alpine:3.20.3

# Add metadata labels
ARG VERSION=dev
ARG BUILD_TIME
ARG VCS_REF
LABEL org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.created="${BUILD_TIME}" \
      org.opencontainers.image.revision="${VCS_REF}" \
      org.opencontainers.image.title="K8s Internal Load Balancer" \
      org.opencontainers.image.description="Kubernetes-native internal load balancer with Traefik" \
      org.opencontainers.image.source="https://github.com/tazhate/k8s-internal-loadbalancer" \
      org.opencontainers.image.licenses="MIT"

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Create a non-root user and group
RUN addgroup -S nonroot && adduser -S nonroot -G nonroot

# Set the working directory inside the container
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/traefik-updater /app/traefik-updater

# Change ownership of the binary to the non-root user
RUN chown nonroot:nonroot /app/traefik-updater

# Switch to the non-root user
USER nonroot

# Expose any necessary ports (if needed)
EXPOSE 3333

# Set the entrypoint to run the binary
CMD ["/app/traefik-updater"]
