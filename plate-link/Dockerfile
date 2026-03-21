FROM golang:1.24-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/plate-link .

FROM alpine:3.21 AS runner
WORKDIR /app

RUN apk add --no-cache ca-certificates wget

ENV HTTP_ADDR=:3600

COPY --from=builder /out/plate-link /app/plate-link

RUN addgroup -S app && adduser -S -G app app
USER app

EXPOSE 3600

ENTRYPOINT ["/app/plate-link"]
