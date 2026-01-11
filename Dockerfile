# Build stage
FROM golang:1.23-alpine AS builder

# Version should match internal/version/version.go
ARG VERSION=1.0.0

WORKDIR /app

# Install git for fetching dependencies
RUN apk add --no-cache git

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o perfowl .

# Runtime stage
FROM alpine:latest

ARG VERSION=1.0.0

# Add ca-certificates for HTTPS and timezone data
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/perfowl /usr/local/bin/perfowl

# Create profiles directory for mounting
RUN mkdir -p /profiles

# Set the entrypoint to MCP mode
ENTRYPOINT ["perfowl", "mcp"]

# Labels for Docker MCP Toolkit
LABEL org.opencontainers.image.title="PerfOwl MCP"
LABEL org.opencontainers.image.description="MCP server for browser performance trace analysis - Optimization Workbench & Lab"
LABEL org.opencontainers.image.version="${VERSION}"
LABEL mcp.tool="true"
