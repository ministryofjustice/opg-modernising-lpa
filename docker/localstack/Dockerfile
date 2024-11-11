FROM golang:1.23.3-alpine AS build

RUN apk add --no-cache zip=3.0-r12 unzip=6.0-r14

WORKDIR /app

COPY --link go.mod go.sum ./
RUN go mod download

COPY --link cmd/event-received ./cmd/event-received
COPY --link cmd/schedule-runner ./cmd/schedule-runner
COPY --link internal ./internal
COPY --link lang ./lang

RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=0 go build -o cmd/event-received/bootstrap ./cmd/event-received \
  && chmod 755 cmd/event-received/bootstrap \
  && zip -j event-received.zip cmd/event-received/bootstrap && zip -r event-received.zip lang

RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=0 go build -o cmd/schedule-runner/bootstrap ./cmd/schedule-runner \
  && chmod 755 cmd/schedule-runner/bootstrap \
  && zip -j schedule-runner.zip cmd/schedule-runner/bootstrap && zip -r schedule-runner.zip lang

FROM --platform=${TARGETARCH} localstack/localstack:3.8.1 AS localstack

COPY --from=build /app/event-received.zip /etc/event-received.zip
COPY --from=build /app/schedule-runner.zip /etc/schedule-runner.zip

COPY ./docker/localstack/localstack-init.sh /etc/localstack/init/ready.d/localstack-init.sh

COPY ./docker/localstack/dynamodb-lpa-gsi-schema.json ./dynamodb-lpa-gsi-schema.json

RUN chmod 544 /etc/localstack/init/ready.d/localstack-init.sh
