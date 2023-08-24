FROM golang:1.21 as build-env

RUN apt-get install -y --no-install-recommends openssl

WORKDIR /app

COPY ./cmd/GOVUKSignIn/main.go .
COPY go.mod go.sum ./

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o signIn main.go

RUN addgroup --system app && \
  adduser --system --gecos app app && \
  chown -R app:app /app

USER app

CMD [ "/app/signIn" ]
