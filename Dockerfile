FROM node:18.7.0-alpine3.16 as asset-env

WORKDIR /app

RUN mkdir -p web/static

COPY app/package.json .
COPY app/yarn.lock .
RUN yarn

COPY web/assets web/assets
RUN yarn build

FROM golang:1.18 as build-env

WORKDIR /app

COPY app/go.mod .

RUN go mod download

COPY /app .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/mlpab

FROM alpine:3.16.1

WORKDIR /go/bin

COPY --from=build-env /go/bin/mlpab /go/bin/mlpab
COPY --from=asset-env /app/web/static web/static
COPY web/template web/template

RUN addgroup -S app && \
    adduser -S -g app app && \
    chown -R app:app mlpab web/template web/static
USER app

ENTRYPOINT ["./mlpab"]
