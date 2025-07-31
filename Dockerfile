FROM golang:1.24-alpine AS builder

RUN apk add --no-cache make

WORKDIR /src
COPY . .
RUN make build

FROM alpine:edge AS certs

RUN apk add --no-cache ca-certificates && update-ca-certificates --fresh

FROM scratch

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /src/bin/seccomp-daemonset /seccomp-daemonset

# Default to 2 cores. We want to teach the runtime to use as little CPU as it can, without crippling it by using only 1 core (which can make goroutines not work in parallel).
# This is a compromise between performance and resource usage, while also reducing from the probable default of 8+ cores on most production machines.
ENV GOMAXPROCS=2

CMD ["/seccomp-daemonset"]
HEALTHCHECK CMD ["/seccomp-daemonset", "healthcheck"]
