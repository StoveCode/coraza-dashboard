# syntax=docker/dockerfile:1

# ---- Build stage ----
FROM golang:1.22-alpine AS builder

WORKDIR /build

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /app/backend ./main.go

# ---- Runtime stage ----
FROM alpine:3.20

# Security updates
RUN apk --no-cache upgrade && apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/backend .

# Non-root user
USER nobody:nobody

EXPOSE 8080

HEALTHCHECK --interval=15s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:8080/api/health || exit 1

CMD ["/app/backend"]
