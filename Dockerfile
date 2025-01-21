# Use an official Golang image for building
FROM golang:1.23.2 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy Go module manifests and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the application
RUN go build -o main .

# Use a minimal image for the runtime
FROM debian:bookworm-slim

# Set the working directory in the container
WORKDIR /app

# Install necessary runtime dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Copy the built application from the builder stage
COPY --from=builder /app/main .

# Copy HTML templates
COPY ./templates ./templates

# Expose the application's port
EXPOSE 8000

# Command to run the application
CMD ["./main"]
