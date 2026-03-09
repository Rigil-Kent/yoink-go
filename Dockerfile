# ── Build stage ────────────────────────────────────────────────────────────
FROM mcr.microsoft.com/oss/go/microsoft/golang:1.22-bullseye AS builder

WORKDIR /app

# Restore modules in a separate layer so it's cached until go.mod/go.sum change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source and build a fully static binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -trimpath -o yoink .

# ── Runtime stage ──────────────────────────────────────────────────────────
# distroless/base-debian12:nonroot — minimal attack surface, non-root by default
FROM gcr.io/distroless/base-debian12:nonroot

LABEL org.opencontainers.image.title="yoink" \
      org.opencontainers.image.description="Comic downloader web UI" \
      org.opencontainers.image.source="https://github.com/bryanlundberg/yoink-go"

WORKDIR /app

COPY --from=builder --chown=nonroot:nonroot /app/yoink .

ENV YOINK_LIBRARY=/library

VOLUME ["/library"]
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD ["/app/yoink", "healthcheck"]

USER nonroot

CMD ["/app/yoink", "serve"]
