# syntax=docker/dockerfile:1

FROM golang:1.27-alpine AS build
# The SQLite driver (mattn/go-sqlite3) uses CGO, so the build needs a C compiler.
RUN apk add --no-cache gcc musl-dev
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go tool templ generate && CGO_ENABLED=1 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM litestream/litestream:0.5.17 AS litestream

FROM alpine:3.22
RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -H -u 10001 app \
    && mkdir -p /data && chown app:app /data
WORKDIR /app
COPY --from=litestream /usr/local/bin/litestream /usr/local/bin/litestream
COPY --from=build /out/server ./server
# The server reads these from its working directory; .importmap lets it
# restore assets/vendor without reaching the CDN.
COPY --chown=app:app assets ./assets
COPY --chown=app:app configurations ./configurations
COPY --chown=app:app .importmap ./.importmap
COPY litestream.yml /etc/litestream.yml
COPY --chmod=755 docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh

USER app
ENV PORT=3000 DATABASE_PATH=/data/app.db
EXPOSE 3000
VOLUME /data
HEALTHCHECK --interval=10s --timeout=3s --start-period=10s --retries=3 \
    CMD wget -qO- "http://127.0.0.1:${PORT}/up" || exit 1
ENTRYPOINT ["docker-entrypoint.sh"]
