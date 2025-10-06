FROM golang:1.25.1-alpine@sha256:b6ed3fd0452c0e9bcdef5597f29cc1418f61672e9d3a2f55bf02e7222c014abd AS base

WORKDIR /app

FROM node:22.20.0-alpine3.21@sha256:b1ab5252ffe4191d7b8728606c5b943fc78d6e306950cef9963e50bcdb3c47a7 AS asset-env

WORKDIR /app

COPY --link package.json pnpm-lock.yaml ./

RUN npm install --global corepack@latest
RUN corepack enable pnpm
RUN pnpm install --prod --frozen-lockfile

COPY --link web/assets web/assets
RUN mkdir -p web/static && pnpm build

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
