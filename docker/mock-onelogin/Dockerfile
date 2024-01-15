FROM golang:1.21.6-alpine as build-env

RUN apk --no-cache add openssl

WORKDIR /app

COPY --link go.mod go.sum ./
RUN go mod download

COPY --link cmd/mock-onelogin ./cmd/mock-onelogin
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -a -installsuffix cgo -o mock-onelogin ./cmd/mock-onelogin

RUN addgroup --system app && \
  adduser --system --gecos app app && \
  chown -R app:app /app

USER app

CMD [ "/app/mock-onelogin" ]
