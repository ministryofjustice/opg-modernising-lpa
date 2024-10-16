FROM golang:1.23.2-alpine AS base

WORKDIR /app

FROM node:20.18.0-alpine3.20 AS asset-env

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

COPY --link go.mod go.sum ./
RUN go mod download

COPY --link cmd/mlpa ./cmd/mlpa
COPY --link internal ./internal

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-X main.Tag=${TAG}" -o /go/bin/mlpab ./cmd/mlpa

FROM alpine:3.20.3 AS production

WORKDIR /go/bin

COPY --link web/robots.txt web/robots.txt
COPY --from=asset-env /app/web/static web/static
COPY --from=build-env /go/bin/mlpab mlpab
COPY --link web/template web/template
COPY --link lang lang

RUN addgroup -S app && \
  adduser -S -g app app && \
  chown -R app:app mlpab web/template web/static web/robots.txt
USER app

ENTRYPOINT ["./mlpab"]
