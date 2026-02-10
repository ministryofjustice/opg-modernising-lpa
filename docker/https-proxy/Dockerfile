FROM golang:1.25.7-alpine@sha256:724e212d86d79b45b7ace725b44ff3b6c2684bfd3131c43d5d60441de151d98e AS base

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
