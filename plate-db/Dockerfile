FROM golang:1.24-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/plate-db .

FROM alpine:3.21 AS runner
WORKDIR /app

ENV HTTP_ADDR=:3500
ENV DATA_DIR=/data

COPY --from=builder /out/plate-db /app/plate-db

RUN addgroup -S app && adduser -S -G app app
USER app

VOLUME ["/data"]
EXPOSE 3500

ENTRYPOINT ["/app/plate-db"]
