# Build stage: Use the official Go image to build the binary.
FROM golang:1.20-alpine AS builder

# Install git if needed to download any dependencies.
RUN apk add --no-cache git

# Set the working directory.
WORKDIR /app

# Copy go.mod and go.sum first; this helps leverage cached layers.
COPY go.mod go.sum ./
RUN go mod download

# Now copy the entire project.
COPY . .

# Build the Go server binary.
# This assumes your main package is in the current directory.
RUN go build -o go-server .

# Run stage: use a lightweight image to run the application.
FROM alpine:latest

# Install certificates (if your app makes HTTPS calls).
RUN apk add --no-cache ca-certificates

# Set the working directory.
WORKDIR /root/

# Copy the binary from the builder stage.
COPY --from=builder /app/go-server .

# Expose port 8080.
EXPOSE 8080

# Run the Go server.
CMD ["./go-server"]
