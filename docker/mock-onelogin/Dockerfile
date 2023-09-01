FROM golang:1.21 as build-env

RUN apt-get install -y --no-install-recommends openssl

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/mock-onelogin ./cmd/mock-onelogin
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o mock-onelogin ./cmd/mock-onelogin

RUN addgroup --system app && \
  adduser --system --gecos app app && \
  chown -R app:app /app

USER app

CMD [ "/app/mock-onelogin" ]
