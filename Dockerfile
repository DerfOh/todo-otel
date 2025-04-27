# Use the official Go image as a base image
FROM golang:1.24.2 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go modules manifests
COPY go.mod go.sum ./

# Download Go modules
RUN go mod download

# Copy the source code
COPY . .

# Build the Go application
RUN go build -o todo-app .

# Use an Alpine image with glibc for the final container
FROM alpine:3.18

# Install glibc
RUN apk add --no-cache libc6-compat

# Set the working directory inside the container
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/todo-app .

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["./todo-app"]