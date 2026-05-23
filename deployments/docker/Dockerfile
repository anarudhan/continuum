# Build stage
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o continuum ./cmd/continuum

# Final stage
FROM alpine:3.21

RUN apk add --no-cache ca-certificates wget

WORKDIR /app
COPY --from=builder /build/continuum /app/continuum

# Copy migration files needed at runtime
COPY --from=builder /build/internal/models/migrations /app/internal/models/migrations

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -q --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["/app/continuum"]
