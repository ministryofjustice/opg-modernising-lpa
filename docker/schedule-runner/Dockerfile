FROM golang:1.23.2-alpine AS build

WORKDIR /app

COPY --link go.mod go.sum ./
RUN go mod download

COPY --link cmd/schedule-runner ./cmd/schedule-runner
COPY --link internal ./internal

RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=0 go build -o schedule-runner ./cmd/schedule-runner

FROM public.ecr.aws/lambda/provided:al2023.2024.10.14.12 AS dev

WORKDIR /app

COPY --from=build /app/schedule-runner /var/task/schedule-runner
COPY --link lang ./lang
COPY --link docker/aws-lambda-rie ./aws-lambda-rie
COPY ./docker/adot-collector/config.yaml /opt/config/config.yaml
COPY ./docker/adot-collector/collector /opt/extensions/collector

## Make the extension executable
RUN chmod 755 /opt/config/config.yaml

# Set environment variables for the collector
ENV AWS_LAMBDA_EXEC_WRAPPER=/opt/otel-handler
ENV OPENTELEMETRY_COLLECTOR_CONFIG_FILE="/opt/config/config.yaml"

ENTRYPOINT ["./schedule-runner"]

FROM public.ecr.aws/lambda/provided:al2023.2024.10.14.12 AS production

WORKDIR /app
COPY --link  docker/install_lambda_insights.sh /app/
COPY --from=otel-ext-layer /opt/ .

RUN chmod +x "/app/install_lambda_insights.sh" \
  && /app/install_lambda_insights.sh "${TARGETPLATFORM}"

COPY --from=build /app/schedule-runner ./schedule-runner
COPY --link lang ./lang

ENTRYPOINT ["./schedule-runner"]
