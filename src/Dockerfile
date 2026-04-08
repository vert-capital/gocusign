# ============================================================================
# STAGE 1: Builder
# ============================================================================
FROM golang:1.25.9-bookworm AS builder

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GO111MODULE=on \
    GOPROXY=https://proxy.golang.org,direct

# Install CA certificates for HTTPS connections
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && apt-get clean

WORKDIR /build

# Copy dependency files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build with security flags (strip debug symbols with -w -s)
RUN go build \
    -ldflags="-w -s" \
    -a \
    -o /build/app_bin .

# Strip debug symbols for smaller binary
RUN strip --strip-unneeded /build/app_bin 2>/dev/null || true

# Create tmp directory (will be copied to final stage)
RUN mkdir -p /tmp && chmod 1777 /tmp

# ============================================================================
# STAGE 2: Final Runtime Image (Distroless)
# ============================================================================
FROM gcr.io/distroless/static-debian12:nonroot

# Security: Run as non-root user (distroless nonroot = UID 65532)
USER nonroot:nonroot

WORKDIR /app

# Copy only the binary from builder
COPY --from=builder --chown=nonroot:nonroot /build/app_bin /app/app_bin

# Copy CA certificates for HTTPS connections
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy tmp directory (distroless has no shell/commands to create it)
COPY --from=builder --chown=nonroot:nonroot /tmp /tmp

EXPOSE 8080

ENTRYPOINT ["/app/app_bin"]
