FROM golang:1.24.3-alpine@sha256:ef18ee7117463ac1055f5a370ed18b8750f01589f13ea0b48642f5792b234044 AS base

WORKDIR /app

FROM node:22.15.0-alpine3.20@sha256:686b8892b69879ef5bfd6047589666933508f9a5451c67320df3070ba0e9807b AS asset-env

WORKDIR /app

COPY --link package.json yarn.lock ./
RUN yarn --prod

COPY --link web/assets web/assets
RUN mkdir -p web/static && yarn build

FROM base AS dev

WORKDIR /app

COPY --link go.mod go.sum ./
RUN go mod download

COPY --link cmd/mlpa ./cmd/mlpa
COPY --link internal ./internal

COPY --from=asset-env /app/web/static web/static

ARG TARGETOS TARGETARCH

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go install github.com/air-verse/air@latest && go install github.com/go-delve/delve/cmd/dlv@latest

ENTRYPOINT ["air"]

FROM base AS build-env

ARG TAG=v0.0.0

WORKDIR /app

RUN addgroup -S user && \
  adduser -S -g user user

COPY --link go.mod go.sum ./
RUN go mod download

COPY --link cmd/mlpa ./cmd/mlpa
COPY --link internal ./internal

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-X main.Tag=${TAG}" -o /go/bin/mlpab ./cmd/mlpa

FROM scratch AS production

WORKDIR /go/bin

COPY --from=asset-env /app/web/static web/static
COPY --from=build-env /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build-env /etc/passwd /etc/passwd
COPY --from=build-env /go/bin/mlpab mlpab
COPY --link web/robots.txt web/robots.txt
COPY --link web/template web/template
COPY --link lang lang

USER user
EXPOSE 8080

ENTRYPOINT ["./mlpab"]
