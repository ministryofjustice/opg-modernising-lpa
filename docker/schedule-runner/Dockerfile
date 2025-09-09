FROM golang:1.25.0-alpine@sha256:f18a072054848d87a8077455f0ac8a25886f2397f88bfdd222d6fafbb5bba440 AS build

WORKDIR /app

COPY --link go.mod go.sum ./
RUN go mod download

COPY --link cmd/schedule-runner ./cmd/schedule-runner
COPY --link internal ./internal

RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=0 go build -ldflags="-X main.Tag=${TAG}" -o schedule-runner ./cmd/schedule-runner

FROM public.ecr.aws/lambda/provided:al2023.2025.09.08.11 AS production

WORKDIR /app

COPY --link  docker/install_lambda_insights.sh /app/

RUN chmod +x "/app/install_lambda_insights.sh" \
  && /app/install_lambda_insights.sh "${TARGETPLATFORM}"

RUN dnf clean all && \
  dnf -y update libxml2 && \
  dnf clean all

# Switch DNF to the latest AL2023.7 release (2023.7.20250512) and update SQLite packages
#Add trivy ignore on this issues as we are updating not installing.
# RUN echo "2023.7.20250512" > /etc/dnf/vars/releasever && \
# dnf clean all && \
# dnf -y update sqlite-libs libxml2 && \
# dnf clean all


COPY --from=build /app/schedule-runner ./schedule-runner
COPY --link lang ./lang
COPY --link ./docker/adot-collector/ /opt

RUN chmod 755 /opt/config/config.yaml

ENV AWS_LAMBDA_EXEC_WRAPPER=/opt/otel-handler
ENV OPENTELEMETRY_COLLECTOR_CONFIG_URI="/opt/config/config.yaml"

ENTRYPOINT ["./schedule-runner"]
