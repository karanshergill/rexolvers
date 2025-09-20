FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies for SQLite
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o rexolvers .

FROM alpine:latest

RUN apk --no-cache add ca-certificates sqlite

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/rexolvers .
COPY --from=builder /app/config.yaml .

# Create directory for database
RUN mkdir -p /app/data

# Set database path to mounted volume
ENV DB_PATH=/app/data/resolvers.db

ENTRYPOINT ["./rexolvers"]