FROM golang:1.25.1-alpine@sha256:b6ed3fd0452c0e9bcdef5597f29cc1418f61672e9d3a2f55bf02e7222c014abd AS build

WORKDIR /app

COPY --link go.mod go.sum ./
RUN go mod download

COPY --link cmd/create-s3-replication-job ./cmd/create-s3-replication-job
COPY --link internal ./internal

RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=0 go build -o create-s3-replication-job ./cmd/create-s3-replication-job

FROM public.ecr.aws/lambda/provided:al2023.2025.08.26.15@sha256:e4f62f62c02c71ee669c7c5feeb2ee8e867aa8b22ef277a6c071ae95d613db4f AS production

WORKDIR /app
COPY --link  docker/install_lambda_insights.sh /app/

RUN chmod +x "/app/install_lambda_insights.sh" \
  && /app/install_lambda_insights.sh "${TARGETPLATFORM}"

# Switch DNF to the latest AL2023.7 release (2023.7.20250512) and update SQLite packages
#Added to trivy ignore on this issues as we are updating not installing.
RUN echo "2023.7.20250512" > /etc/dnf/vars/releasever && \
  dnf clean all && \
  dnf -y update sqlite-libs libxml2 && \
  dnf clean all

COPY --from=build /app/create-s3-replication-job ./create-s3-replication-job
COPY --link lang ./lang
COPY --link ./docker/adot-collector/ /opt

RUN chmod 755 /opt/config/config.yaml

ENV AWS_LAMBDA_EXEC_WRAPPER=/opt/otel-handler
ENV OPENTELEMETRY_COLLECTOR_CONFIG_URI="/opt/config/config.yaml"

ENTRYPOINT ["./create-s3-replication-job"]
