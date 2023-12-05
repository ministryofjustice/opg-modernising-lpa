FROM golang:1.21.4-alpine as base

ARG ARCH=

WORKDIR /app

FROM node:18.16.0-alpine3.16 as asset-env

WORKDIR /app

COPY package.json yarn.lock ./
RUN yarn --prod

COPY web/assets web/assets
RUN mkdir -p web/static && yarn build

FROM base AS dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/mlpa ./cmd/mlpa
COPY internal ./internal

COPY --from=asset-env /app/web/static web/static

RUN CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} go install github.com/cosmtrek/air@latest && go install github.com/go-delve/delve/cmd/dlv@latest

ENTRYPOINT ["air"]

FROM base as build-env

ARG TAG=v0.0.0

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/mlpa ./cmd/mlpa
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} go build -ldflags="-X main.Tag=${TAG}" -o /go/bin/mlpab ./cmd/mlpa

FROM alpine:3.18.5 as production

WORKDIR /go/bin

COPY web/robots.txt web/robots.txt
COPY --from=asset-env /app/web/static web/static
COPY --from=build-env /go/bin/mlpab mlpab
COPY web/template web/template
COPY lang lang

RUN addgroup -S app && \
  adduser -S -g app app && \
  chown -R app:app mlpab web/template web/static web/robots.txt
USER app

ENTRYPOINT ["./mlpab"]
