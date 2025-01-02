FROM golang:1.23.4-alpine AS build

WORKDIR /app

COPY --link go.mod go.sum ./
RUN go mod download

COPY --link cmd/create-s3-replication-job ./cmd/create-s3-replication-job
COPY --link internal ./internal

RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=0 go build -o create-s3-replication-job ./cmd/create-s3-replication-job

FROM public.ecr.aws/lambda/provided:al2023.2024.10.14.12 AS dev

WORKDIR /app

COPY --from=build /app/create-s3-replication-job /var/task/create-s3-replication-job
COPY --link lang ./lang
COPY --link ./docker/adot-collector/ /opt
COPY --link docker/aws-lambda-rie ./aws-lambda-rie

ENV AWS_LAMBDA_EXEC_WRAPPER=/opt/otel-handler
ENV OPENTELEMETRY_COLLECTOR_CONFIG_FILE="/opt/config/config.yaml"

ENTRYPOINT ["./create-s3-replication-job"]

FROM public.ecr.aws/lambda/provided:al2023.2024.10.14.12 AS production

WORKDIR /app
COPY --link  docker/install_lambda_insights.sh /app/

RUN chmod +x "/app/install_lambda_insights.sh" \
  && /app/install_lambda_insights.sh "${TARGETPLATFORM}"

COPY --from=build /app/create-s3-replication-job ./create-s3-replication-job
COPY --link lang ./lang
COPY --link ./docker/adot-collector/ /opt

RUN chmod 755 /opt/config/config.yaml

ENV AWS_LAMBDA_EXEC_WRAPPER=/opt/otel-handler
ENV OPENTELEMETRY_COLLECTOR_CONFIG_FILE="/opt/config/config.yaml"

ENTRYPOINT ["./create-s3-replication-job"]
