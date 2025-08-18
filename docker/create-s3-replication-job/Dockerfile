FROM golang:1.25.0-alpine@sha256:5dffbd69b37b8e9fc35d94817e94bc7bf4fd047147e3c5353d357bd3492992bf AS build

WORKDIR /app

COPY --link go.mod go.sum ./
RUN go mod download

COPY --link cmd/create-s3-replication-job ./cmd/create-s3-replication-job
COPY --link internal ./internal

RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=0 go build -o create-s3-replication-job ./cmd/create-s3-replication-job

FROM public.ecr.aws/lambda/provided:al2023.2025.08.05.12@sha256:a6fd59cc309cef2fb8585bb15f60d74a81bedbe2f46cf2a73a4307513925c884 AS production

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
