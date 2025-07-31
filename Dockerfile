FROM golang:1.24-alpine AS builder

RUN apk add --no-cache make

WORKDIR /src
COPY . .
RUN make build

FROM alpine:edge AS certs

RUN apk add --no-cache ca-certificates && update-ca-certificates --fresh

FROM scratch

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /src/bin/seccomp-operator /seccomp-operator

CMD ["/seccomp-operator"]
HEALTHCHECK CMD ["/seccomp-operator", "healthcheck"]
