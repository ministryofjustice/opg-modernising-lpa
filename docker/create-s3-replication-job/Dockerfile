FROM golang:1.23.5-alpine AS adot-builder-alpine
WORKDIR /app
RUN apk add --no-cache\
  git=2.47.2-r0 \
  bash=5.2.37-r0 \
  make=4.4.1-r2 \
  zip=3.0-r13 \
  unzip=6.0-r15 && \
  git clone --recurse-submodules https://github.com/aws-observability/aws-otel-lambda.git
WORKDIR /app/aws-otel-lambda/
RUN git checkout -b newbranch f6d6301
WORKDIR /app/aws-otel-lambda/go
RUN ./build.sh

FROM golang:1.23.5-alpine AS build

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
COPY --link ./docker/adot-collector/config/ /opt/collector-config/config.yaml
COPY --from=adot-builder-alpine /app/aws-otel-lambda/opentelemetry-lambda/collector/build/extensions/collector /opt/extensions/collector
COPY --link docker/aws-lambda-rie ./aws-lambda-rie

ENV AWS_LAMBDA_EXEC_WRAPPER=/opt/otel-handler
ENV OPENTELEMETRY_COLLECTOR_CONFIG_FILE="/opt/collector-config/config.yaml"

ENTRYPOINT ["./create-s3-replication-job"]

FROM public.ecr.aws/lambda/provided:al2023.2024.10.14.12 AS production

WORKDIR /app
COPY --link  docker/install_lambda_insights.sh /app/

RUN chmod +x "/app/install_lambda_insights.sh" \
  && /app/install_lambda_insights.sh "${TARGETPLATFORM}"

COPY --from=build /app/create-s3-replication-job ./create-s3-replication-job
COPY --link lang ./lang
COPY --link ./docker/adot-collector/config/ /opt/collector-config/config.yaml
COPY --from=adot-builder-alpine /app/aws-otel-lambda/opentelemetry-lambda/collector/build/extensions/collector /opt/extensions/collector

RUN chmod 755 /opt/collector-config/config.yaml

ENV AWS_LAMBDA_EXEC_WRAPPER=/opt/otel-handler
ENV OPENTELEMETRY_COLLECTOR_CONFIG_FILE="/opt/collector-config/config.yaml"

ENTRYPOINT ["./create-s3-replication-job"]
