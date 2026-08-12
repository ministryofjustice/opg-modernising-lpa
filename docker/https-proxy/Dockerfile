FROM golang:1.26.5-alpine@sha256:0178a641fbb4858c5f1b48e34bdaabe0350a330a1b1149aabd498d0699ff5fb2 AS base

ARG ARCH=amd64

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/https-proxy ./cmd/https-proxy

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /go/bin/https-proxy ./cmd/https-proxy

FROM scratch AS production

WORKDIR /app

COPY --from=base /go/bin/https-proxy https-proxy
COPY --link docker/https-proxy/cert.pem docker/https-proxy/key.pem .

ENTRYPOINT ["./https-proxy"]
