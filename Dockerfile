# Build
FROM golang:1.24-alpine AS build

WORKDIR /src
# Copy the module files first so dependency resolution is cached separately
# from the source. With no external dependencies this is nearly free, but it
# keeps the layer order correct if that ever changes.
COPY go.mod ./
RUN go mod download

COPY . .
ARG VERSION=docker
RUN CGO_ENABLED=0 go build -trimpath \
    -ldflags "-s -w -X main.version=${VERSION}" \
    -o /out/ptal ./cmd/ptal

# Run
FROM alpine:3.20

# TLS roots are required to reach api.github.com and api.telegram.org; tzdata
# makes QUIET_HOURS respect the TZ environment variable instead of assuming UTC.
RUN apk add --no-cache ca-certificates tzdata \
 && adduser -D -u 10001 ptal

COPY --from=build /out/ptal /usr/local/bin/ptal

# State lives here so it survives container replacement. Mount a volume.
ENV STATE_PATH=/data/state.json
RUN mkdir -p /data && chown ptal:ptal /data
VOLUME /data

USER ptal
WORKDIR /data

# `run` in the foreground: the container runtime is the supervisor here, so
# the built-in service installer is not used.
ENTRYPOINT ["/usr/local/bin/ptal"]
CMD ["run"]
