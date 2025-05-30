FROM golang:1.24.2-alpine3.21 AS builder

WORKDIR /app
COPY . /app
RUN CGO_ENABLED=0 go build .

FROM alpine:3.21

WORKDIR /app

# Install ca-certificates and wget for HTTPS and healthcheck
RUN apk add --no-cache ca-certificates wget

COPY --from=builder /app/glance .

HEALTHCHECK --timeout=10s --start-period=60s --interval=60s \
  CMD wget --spider -q http://localhost:8080/api/healthz

EXPOSE 8080/tcp
ENTRYPOINT ["/app/glance", "--config", "/app/config/glance.yml"]

ENV GITHUB_TOKEN=${GITHUB_TOKEN}


