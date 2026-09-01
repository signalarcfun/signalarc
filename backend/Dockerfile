# SignalArc Backend - Local Development / Cloud Run Dockerfile
# Multi-stage build for a small, maintainable Go image.
#
# Runtime stage also installs Node.js, npm, and the global @circle-fin/cli
# package so the backend can invoke the Circle Agent Wallet CLI in the same
# way as the local WSL development environment.
#
# This image must NOT bake any Circle credentials, OTP request IDs, session
# files, private keys, or API keys. Circle CLI session strategy in production
# remains an unsolved runtime concern handled outside the image.

FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /signalarc-api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /analytics-backfill ./cmd/analytics-backfill

FROM node:20.18.2-alpine3.20

# Runtime dependencies:
# - ca-certificates: required for outbound HTTPS calls
# - node/npm are provided by the pinned Node base image.
#   Keep Node >=20.18.2 because @circle-fin/cli@0.0.3 declares that
#   engine requirement and Circle CLI is a backend runtime dependency.
RUN apk add --no-cache \
        ca-certificates \
    && npm install -g @circle-fin/cli \
    && npm cache clean --force

COPY --from=builder /signalarc-api /usr/local/bin/signalarc-api
COPY --from=builder /analytics-backfill /usr/local/bin/analytics-backfill

EXPOSE 4000

ENTRYPOINT ["signalarc-api"]
