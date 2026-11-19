# syntax=docker/dockerfile:1

ARG GO_VERSION=1.27.4

FROM golang:${GO_VERSION}-alpine AS dev

RUN apk add --no-cache zip \
    && addgroup -g 1000 remotehand \
    && adduser -D -u 1000 -G remotehand remotehand \
    && mkdir -p /workspace /home/remotehand/go/pkg/mod /home/remotehand/.cache/go-build \
    && chown -R remotehand:remotehand /workspace /home/remotehand

ENV GOCACHE=/home/remotehand/.cache/go-build
ENV GOPATH=/home/remotehand/go
WORKDIR /workspace
USER remotehand

CMD ["go", "test", "./..."]

FROM dev AS server-build

COPY --chown=remotehand:remotehand . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /tmp/remotehand \
    ./cmd/remotehand

FROM scratch AS server

COPY --from=server-build /tmp/remotehand /remotehand
EXPOSE 18080
USER 1000:1000
ENTRYPOINT ["/remotehand"]
