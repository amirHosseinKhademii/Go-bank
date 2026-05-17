# syntax=docker/dockerfile:1

# ---- Builder ----
FROM golang:1.25-alpine AS builder
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Cache go.mod and go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the binary
RUN go build -ldflags="-s -w" -o /bank ./cmd/main.go

# ---- Runtime ----
FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /bank /app/bank

# Expose the default port (if applicable, adjust as needed)
EXPOSE 8080

# Command to run the binary
CMD ["/app/bank"]
