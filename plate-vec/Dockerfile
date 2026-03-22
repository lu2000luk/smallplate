FROM golang:1.24-alpine AS builder
WORKDIR /src

ARG TARGETOS=linux
ARG TARGETARCH

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" -o /out/plate-vec .

FROM alpine:3.21 AS runner
WORKDIR /app

RUN apk add --no-cache ca-certificates wget

ENV HTTP_ADDR=:3300
ENV DATA_DIR=/data
ENV CHROMA_URL=http://vec-chroma:8000

COPY --from=builder /out/plate-vec /app/plate-vec

RUN addgroup -S app && adduser -S -G app app
USER app

VOLUME ["/data"]
EXPOSE 3300

ENTRYPOINT ["/app/plate-vec"]
