# ---- Build Stage ----
# Use an official Golang image that includes the Go toolchain.
FROM golang:1.23.5-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum to leverage Docker's layer caching.
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code into the container
COPY . .

# Build the Go application.
# -a: force rebuilding of packages that are already up-to-date.
# -mod=readonly: disallows the go command from automatically updating go.mod and go.sum
RUN CGO_ENABLED=0 GOOS=linux go build -a -mod=readonly -v -o v3-notifications-service .

# ---- Final Stage ----
# Use a minimal alpine image for the final container.
FROM alpine:latest

# Add ca-certificates to allow for secure TLS/SSL connections to other services.
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the pre-built binary from the "builder" stage.
COPY --from=builder /app/v3-notifications-service .

# Command to run the executable when the container starts.
# Cloud Run will use the PORT environment variable, which your app reads.
EXPOSE 8080
CMD ["./v3-notifications-service"]