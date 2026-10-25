# zcode-gateway (Go) —— 无 cgo 纯静态二进制，运行层不需要 Chromium
FROM golang:1.27-bookworm AS build
WORKDIR /src
COPY src/ ./
ENV CGO_ENABLED=0 \
    GOFLAGS=-mod=mod
RUN go build -trimpath -ldflags "-s -w" -o /out/zcode-gateway . && /out/zcode-gateway -h 2>/dev/null || true

FROM debian:bookworm-slim
RUN apt-get update \
 && apt-get install -y --no-install-recommends ca-certificates tzdata curl \
 && rm -rf /var/lib/apt/lists/*
ENV TZ=Asia/Shanghai \
    HOME=/app \
    ZCODE_GATEWAY_ADDR=0.0.0.0:8787 \
    ZCODE_GATEWAY_DATA=/app/data
WORKDIR /app
COPY --from=build /out/zcode-gateway /app/zcode-gateway
RUN groupadd -g 10001 app && useradd -u 10001 -g 10001 -r -s /usr/sbin/nologin app \
 && mkdir -p /app/data && chown -R 10001:10001 /app
USER 10001:10001
EXPOSE 8787
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD curl -fsS http://127.0.0.1:8787/health || exit 1
ENTRYPOINT ["/app/zcode-gateway"]
