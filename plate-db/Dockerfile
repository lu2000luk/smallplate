FROM golang:1.24-bookworm AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/plate-db .

FROM debian:bookworm-slim AS runner
WORKDIR /app

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates wget \
    && rm -rf /var/lib/apt/lists/*

ENV HTTP_ADDR=:3500
ENV DATA_DIR=/data

COPY --from=builder /out/plate-db /app/plate-db

VOLUME ["/data"]
EXPOSE 3500

ENTRYPOINT ["/app/plate-db"]
