FROM golang:1.23.2-alpine AS build-env

RUN apk --no-cache add openssl

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/mock-notify ./cmd/mock-notify

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -a -installsuffix cgo -o mock-notify ./cmd/mock-notify

RUN addgroup --system app && \
  adduser --system --gecos app app && \
  chown -R app:app /app

USER app

CMD [ "/app/mock-notify" ]
