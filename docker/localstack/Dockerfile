FROM golang:1.23.2-alpine AS build

RUN apk add zip

WORKDIR /app

COPY --link go.mod go.sum ./
RUN go mod download

COPY --link cmd/event-received ./cmd/event-received
COPY --link internal ./internal

RUN GOOS=linux GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=0 go build -tags lambda.norpc -o event-received ./cmd/event-received
RUN zip event-received.zip event-received

FROM localstack/localstack:3.8.0 AS localstack

COPY --from=build /app/event-received.zip /etc/event-received.zip

COPY ./docker/localstack/localstack-init.sh /etc/localstack/init/ready.d/localstack-init.sh
COPY ./docker/localstack/dynamodb-lpa-gsi-schema.json ./dynamodb-lpa-gsi-schema.json

RUN chmod 544 /etc/localstack/init/ready.d/localstack-init.sh
