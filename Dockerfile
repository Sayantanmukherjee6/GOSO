# Stage 1: Build the application
FROM golang:1.23.2 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go modules manifests and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire source code to the container
COPY . .

# Build the Go application
RUN go build -o main .

# Stage 2: Create a lightweight final image
FROM debian:bullseye-slim

# Set the working directory
WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/main .

# Copy the HTML templates to the container
COPY ./template ./template

# Set the default port used by the application
EXPOSE 8000

# Command to run the application
CMD ["./main"]
