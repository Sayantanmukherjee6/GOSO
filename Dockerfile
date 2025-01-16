# Use the official Go image as a base image
FROM golang:1.20-alpine

# Set environment variables for Go
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Create and set the working directory in the container
WORKDIR /app

# Copy Go modules manifest files
COPY go.mod go.sum ./

# Download Go module dependencies
RUN go mod download

# Copy the source code to the container
COPY *.go ./
COPY templates/ ./templates/

# Build the Go application
RUN go build -o schat .

# Expose the port on which the app runs
EXPOSE 8000

# Command to run the application
CMD ["./goso"]
