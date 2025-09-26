FROM golang:1.25.1-alpine@sha256:b6ed3fd0452c0e9bcdef5597f29cc1418f61672e9d3a2f55bf02e7222c014abd AS base

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
