# ============================================================================
# STAGE 1: Builder
# ============================================================================
FROM golang:1.25.9-bookworm AS builder

ARG PROJECT_NAME="app"
ARG VERSION="dev"
ARG COMMIT="none"

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GOPROXY=https://proxy.golang.org,direct

WORKDIR /build

# Copy dependency files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN go build -ldflags="-w -s" \
    -a -installsuffix cgo \
    -o /build/app_bin .

# ============================================================================
# STAGE 2: Final Runtime Image (Distroless)
# ============================================================================
# Static binary — no glibc needed, using static-debian12 avoids glibc CVEs
FROM gcr.io/distroless/static-debian12:nonroot

USER nonroot:nonroot

WORKDIR /app

COPY --from=builder --chown=nonroot:nonroot /build/app_bin /app/app_bin
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 8080

ENTRYPOINT ["/app/app_bin"]
