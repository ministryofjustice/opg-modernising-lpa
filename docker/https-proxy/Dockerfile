FROM golang:1.25.6-alpine@sha256:98e6cffc31ccc44c7c15d83df1d69891efee8115a5bb7ede2bf30a38af3e3c92 AS base

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
