FROM golang:1.24.5-alpine@sha256:daae04ebad0c21149979cd8e9db38f565ecefd8547cf4a591240dc1972cf1399 AS base

WORKDIR /app

FROM node:22.17.1-alpine3.21@sha256:f00440c423ce5657e4d2347fa3b9bf5887389f6fcf3daa25cc79ea11a6a00f80 AS asset-env

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
