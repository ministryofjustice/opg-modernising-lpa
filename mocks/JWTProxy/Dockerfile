FROM golang:1.18 as build-env

RUN apt-get install openssl

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o proxy main.go

RUN openssl genpkey -algorithm RSA -out ./private_key.pem -pkeyopt rsa_keygen_bits:2048 && \
  openssl rsa -pubout -in ./private_key.pem -out ./public_key.pem

RUN addgroup --system app && \
    adduser --system --gecos app app && \
    chown -R app:app /app

USER app

#ENTRYPOINT [ "/app/entrypoint.sh" ]

CMD [ "/app/proxy" ]
