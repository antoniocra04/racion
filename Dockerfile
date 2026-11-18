# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS build
WORKDIR /src
ENV GOTOOLCHAIN=local

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -mod=readonly -trimpath -ldflags="-s -w" -o /out/ration ./cmd/ration

FROM alpine:3.23
RUN mkdir /data && touch /data/.keep && chown -R 10001:10001 /data && chmod 0700 /data
COPY --from=build /out/ration /usr/local/bin/ration

USER 10001:10001
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/ration"]
CMD ["-addr", "0.0.0.0:8080", "-db", "/data/ration.db"]
