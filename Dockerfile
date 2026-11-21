# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM node:24-alpine AS frontend
WORKDIR /src/frontend
COPY frontend/package.json frontend/package-lock.json ./
COPY frontend/vendor ./vendor
RUN npm ci --no-audit --no-fund
COPY frontend/index.html frontend/tsconfig.json frontend/vite.config.ts ./
COPY frontend/src ./src
RUN npm run build

FROM --platform=$BUILDPLATFORM golang:1.26.11-alpine AS build
WORKDIR /src
ENV GOTOOLCHAIN=local

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
COPY --from=frontend /src/internal/web/spa/generated ./internal/web/spa/generated
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -mod=readonly -trimpath -ldflags="-s -w" -o /out/ration ./cmd/ration

FROM alpine:3.23
RUN mkdir /data && touch /data/.keep && chown -R 10001:10001 /data && chmod 0700 /data
COPY --from=build /out/ration /usr/local/bin/ration

USER 10001:10001
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s CMD wget -qO /dev/null http://127.0.0.1:8080/healthz || exit 1
ENTRYPOINT ["/usr/local/bin/ration"]
CMD ["-addr", "0.0.0.0:8080", "-db", "/data/ration.db"]
