FROM golang:1.23.2-alpine AS build-env

RUN apk --no-cache add openssl=3.3.2-r0

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/mock-os-api ./cmd/mock-os-api
COPY docker/mock-os-api/data ./data

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -a -installsuffix cgo -o mock-os-api ./cmd/mock-os-api

RUN addgroup --system app && \
  adduser --system --gecos app app && \
  chown -R app:app /app

USER app

CMD [ "/app/mock-os-api" ]
