# FROM golang:1.23.5-alpine AS adot-builder-alpine-1
# WORKDIR /app
# RUN apk add --no-cache \
#   git=2.47.2-r0 \
#   bash=5.2.37-r0 \
#   make=4.4.1-r2 \
#   zip=3.0-r13 \
#   unzip=6.0-r15 && \
#   git clone https://github.com/aws-observability/aws-otel-collector
# WORKDIR /app/aws-otel-collector/
# RUN git checkout -b newbranch tags/v0.40.0
# RUN make amd64-build-only
# FROM golang:1.23.5-alpine AS adot-builder-alpine-2
# WORKDIR /app
# RUN apk add --no-cache \
#   git=2.47.2-r0 \
#   bash=5.2.37-r0 \
#   make=4.4.1-r2 \
#   zip=3.0-r13 \
#   unzip=6.0-r15 && \
#   git clone https://github.com/open-telemetry/opentelemetry-lambda
# WORKDIR /app/opentelemetry-lambda/
# # commit hash 21052d8 is 0.36.0
# RUN git checkout -b newbranch 21052d8
# WORKDIR /app/opentelemetry-lambda/collector
# RUN make package
FROM golang:1.23.5-alpine AS adot-builder-alpine
WORKDIR /app
RUN apk add --no-cache \
git=2.47.2-r0 \
bash=5.2.37-r0 \
make=4.4.1-r2 \
zip=3.0-r13 \
unzip=6.0-r15 && \
git clone --recurse-submodules https://github.com/aws-observability/aws-otel-lambda
WORKDIR /app/aws-otel-lambda/
# commit hash 84e2c7fe8444c7476cf7eb6d97c009dac4da1615 is collector layer v0.42.0
# commit hash b00475323dc3f1964151770dbea82d308d6924e3 is collector layer v0.36.0
RUN git checkout -b newbranch 84e2c7fe8444c7476cf7eb6d97c009dac4da1615
WORKDIR /app/aws-otel-lambda/opentelemetry-lambda/collector
RUN make package

FROM golang:1.23.5-alpine AS build

WORKDIR /app

COPY --link go.mod go.sum ./
RUN go mod download

COPY --link cmd/create-s3-replication-job ./cmd/create-s3-replication-job
COPY --link internal ./internal

RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=0 go build -o create-s3-replication-job ./cmd/create-s3-replication-job

FROM public.ecr.aws/lambda/provided:al2023.2024.10.14.12 AS production

WORKDIR /app
COPY --link  docker/install_lambda_insights.sh /app/

RUN chmod +x "/app/install_lambda_insights.sh" \
  && /app/install_lambda_insights.sh "${TARGETPLATFORM}"

COPY --from=build /app/create-s3-replication-job ./create-s3-replication-job
COPY --link ./docker/adot-collector/config/config.yaml /opt/collector-config/config.yaml
#1
# COPY --from=adot-builder-alpine /app/aws-otel-collector/build/linux/amd64/aoc /opt/extensions/collector
# 2
# COPY --from=adot-builder-alpine /app/opentelemetry-lambda/collector/build/extensions/collector /opt/extensions/collector
# 3
COPY --from=adot-builder-alpine /app/aws-otel-lambda/opentelemetry-lambda/collector/build/extensions/collector /opt/extensions/collector

RUN chmod 755 /opt/collector-config/config.yaml

ENV AWS_LAMBDA_EXEC_WRAPPER=/opt/otel-handler
# ENV OPENTELEMETRY_COLLECTOR_CONFIG_URI="/opt/collector-config/config.yaml"

ENTRYPOINT ["./create-s3-replication-job"]
