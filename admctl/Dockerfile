# Build stage
FROM golang:1.23-alpine3.20 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the application
# RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o devenv-mesh-controller .
RUN go build -o devenv-mesh-controller .

# Final stage
FROM alpine:3.20

# Install ca-certificates
RUN apk --no-cache add ca-certificates

# Set the working directory
WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/devenv-mesh-controller .

# Create a directory for the TLS certificates
RUN mkdir -p /etc/webhook/certs

# Expose the port the app runs on
EXPOSE 8443

# Run the binary
CMD ["./devenv-mesh-controller"]
