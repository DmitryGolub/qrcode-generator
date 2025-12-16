FROM golang:1.25-bookworm AS builder

WORKDIR /build

COPY backend/go.mod backend/go.sum ./backend/
COPY backend/pkg ./backend/pkg

COPY gateway/go.mod gateway/go.sum ./gateway/

WORKDIR /build/gateway
RUN go mod download

COPY gateway .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o gateway ./cmd/main.go

FROM debian:bullseye-slim

RUN groupadd -r appuser && useradd -r -g appuser appuser
WORKDIR /app

COPY --chown=appuser:appuser --from=builder /build/gateway/gateway .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

USER appuser

EXPOSE 8080

CMD ["./gateway"]