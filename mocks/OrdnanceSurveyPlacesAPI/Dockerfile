FROM golang:1.20 as build-env

RUN apt-get install -y --no-install-recommends openssl

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o ordnance-survey-places-api main.go

RUN addgroup --system app && \
  adduser --system --gecos app app && \
  chown -R app:app /app

USER app

CMD [ "/app/ordnance-survey-places-api" ]
