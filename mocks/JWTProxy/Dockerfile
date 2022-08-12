FROM golang:1.18-alpine3.16 as build-env

RUN apk add openssl

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o proxy main.go

RUN addgroup -S app && \
    adduser -S -g app app && \
    chown -R app:app /app/entrypoint.sh /app/proxy

USER app

#RUN openssl genpkey -algorithm RSA -out private_key.pem -pkeyopt rsa_keygen_bits:2048 && \
#  openssl rsa -pubout -in private_key.pem -out public_key.pem

#ENTRYPOINT [ "/app/entrypoint.sh" ]

CMD [ "/app/proxy" ]

EXPOSE 5000
