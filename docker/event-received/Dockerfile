FROM golang:1.24.3-alpine@sha256:ef18ee7117463ac1055f5a370ed18b8750f01589f13ea0b48642f5792b234044 AS build

WORKDIR /app

COPY --link go.mod go.sum ./
RUN go mod download

COPY --link cmd/event-received ./cmd/event-received
COPY --link internal ./internal

RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=0 go build -o event-received ./cmd/event-received

FROM public.ecr.aws/lambda/provided:al2023.2025.05.04.04@sha256:fed9eb1f995d9c1f714794e3c2223fd5a97990022eedbab6f6f0d711ba888ac6 AS production

WORKDIR /app
COPY --link  docker/install_lambda_insights.sh /app/

RUN chmod +x /app/install_lambda_insights.sh \
  && /app/install_lambda_insights.sh "${TARGETPLATFORM}"

# Switch DNF to the latest AL2023.7 release (2023.7.20250512) and update SQLite packages
#Add trivy ignore on this issues as we are updating not installing.
RUN echo "2023.7.20250512" > /etc/dnf/vars/releasever && \
dnf clean all && \
dnf -y update sqlite-libs libxml2 && \
dnf clean all

COPY --from=build /app/event-received ./event-received
COPY --link lang ./lang
COPY --link ./docker/adot-collector/ /opt

RUN chmod 755 /opt/config/config.yaml

ENV AWS_LAMBDA_EXEC_WRAPPER=/opt/otel-handler
ENV OPENTELEMETRY_COLLECTOR_CONFIG_URI="/opt/config/config.yaml"

ENTRYPOINT ["./event-received"]
