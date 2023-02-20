FROM node:18.14.0-alpine3.16 as asset-env

WORKDIR /app

COPY package.json yarn.lock ./
RUN yarn --prod

COPY web/assets web/assets
RUN mkdir -p web/static && yarn build

FROM golang:1.20 as build-env

WORKDIR /app
ARG TAG=v0.0.0

COPY app/go.mod app/go.sum ./
RUN go mod download

COPY /app .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-X main.Tag=${TAG}" -o /go/bin/mlpab

FROM alpine:3.17.2 as production

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
