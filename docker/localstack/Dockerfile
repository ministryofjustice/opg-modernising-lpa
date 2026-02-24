FROM golang:1.25.7-alpine@sha256:f6751d823c26342f9506c03797d2527668d095b0a15f1862cddb4d927a7a4ced AS build

RUN apk add --no-cache zip=3.0-r13 unzip=6.0-r16

WORKDIR /app

COPY --link go.mod go.sum ./
RUN go mod download

COPY --link cmd/event-received ./cmd/event-received
COPY --link cmd/schedule-runner ./cmd/schedule-runner
COPY --link cmd/scheduled-task-adder ./cmd/scheduled-task-adder
COPY --link internal ./internal
COPY --link lang ./lang

RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=0 go build -o cmd/event-received/bootstrap ./cmd/event-received \
  && chmod 755 cmd/event-received/bootstrap \
  && zip -j event-received.zip cmd/event-received/bootstrap && zip -r event-received.zip lang

RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=0 go build -ldflags="-X main.Tag=${TAG}" -o cmd/schedule-runner/bootstrap ./cmd/schedule-runner \
  && chmod 755 cmd/schedule-runner/bootstrap \
  && zip -j schedule-runner.zip cmd/schedule-runner/bootstrap && zip -r schedule-runner.zip lang

RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=0 go build -o cmd/scheduled-task-adder/bootstrap ./cmd/scheduled-task-adder \
  && chmod 755 cmd/scheduled-task-adder/bootstrap \
  && zip -r -j scheduled-task-adder.zip cmd/scheduled-task-adder/bootstrap

FROM localstack/localstack:4.12.0@sha256:0df3a97da57de03a588c05d9b8f390f15c7033fc7c4512f94619d344bc3cd317 AS localstack

COPY --from=build /app/event-received.zip /etc/event-received.zip
COPY --from=build /app/schedule-runner.zip /etc/schedule-runner.zip
COPY --from=build /app/scheduled-task-adder.zip /etc/scheduled-task-adder.zip

COPY ./docker/localstack/localstack-init.sh /etc/localstack/init/ready.d/localstack-init.sh
COPY ./docker/localstack/dynamodb-lpa-gsi-schema.json /usr/dynamodb-lpa-gsi-schema.json
COPY ./docker/localstack/schedule-runner-metrics-query.json /usr/schedule-runner-metrics-query.json

RUN chmod 544 /etc/localstack/init/ready.d/localstack-init.sh
