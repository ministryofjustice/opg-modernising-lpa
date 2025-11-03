FROM golang:1.25.3-alpine@sha256:aee43c3ccbf24fdffb7295693b6e33b21e01baec1b2a55acc351fde345e9ec34 AS base

WORKDIR /app

FROM node:22.21.1-alpine3.21@sha256:af8023ec879993821f6d5b21382ed915622a1b0f1cc03dbeb6804afaf01f8885 AS asset-env

WORKDIR /app

COPY --link package.json package-lock.json ./

RUN npm install --omit=optional --omit=dev --frozen-lockfile

COPY --link web/assets web/assets
RUN mkdir -p web/static && npm run build

FROM base AS dev

WORKDIR /app

COPY --link go.mod go.sum ./
RUN go mod download

COPY --link cmd/mlpa ./cmd/mlpa
COPY --link internal ./internal

COPY --from=asset-env /app/web/static web/static

ARG TARGETOS TARGETARCH

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go install github.com/air-verse/air@v1.63.0 && go install github.com/go-delve/delve/cmd/dlv@v1.25.2

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
