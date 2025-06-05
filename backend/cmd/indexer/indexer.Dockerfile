# syntax=docker/dockerfile:1

FROM golang:1.24.3-alpine

WORKDIR /app

# Install git
RUN apk add --no-cache git

# Copy Go modules
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copies entire backend folder
COPY backend/ ./

# Build binary
RUN go build -o indexer ./cmd/indexer

CMD ["./indexer"]