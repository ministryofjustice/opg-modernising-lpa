FROM golang:1.19 as build-env

RUN apt-get install -y --no-install-recommends openssl

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o govuk-notify main.go

RUN addgroup --system app && \
  adduser --system --gecos app app && \
  chown -R app:app /app

USER app

CMD [ "/app/govuk-notify" ]
