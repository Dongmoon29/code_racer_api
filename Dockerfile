# Stage 1: Build the binary
FROM golang:1.22-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the working Directory inside the container
COPY . .

# Set the working directory to the location of main.go
WORKDIR /app/cmd/api

# Build the Go app and place the binary in /app
RUN go build -o /app/main .

# Stage 2: Run the binary
FROM alpine:latest

WORKDIR /app

# Copy the binary from the build stage
COPY --from=builder /app/main .

# Command to run the executable
CMD ["./main"]