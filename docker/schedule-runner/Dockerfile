FROM golang:1.26.0-alpine@sha256:d4c4845f5d60c6a974c6000ce58ae079328d03ab7f721a0734277e69905473e5 AS build

WORKDIR /app

COPY --link go.mod go.sum ./
RUN go mod download

COPY --link cmd/schedule-runner ./cmd/schedule-runner
COPY --link internal ./internal

RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=0 go build -ldflags="-X main.Tag=${TAG}" -o schedule-runner ./cmd/schedule-runner

FROM public.ecr.aws/lambda/provided:al2023.2026.02.18.07@sha256:059ed067a8e796d891100888928e32cde75d63d459617260a7901e217a7249d6 AS production

WORKDIR /app

COPY --link  docker/install_lambda_insights.sh /app/

RUN chmod +x "/app/install_lambda_insights.sh" \
  && /app/install_lambda_insights.sh "${TARGETPLATFORM}"

# Switch DNF to the latest AL2023.10 release (2023.10.20260202) and update SQLite packages
# Add trivy ignore on this issues as we are updating not installing.
RUN echo "2023.10.20260202" > /etc/dnf/vars/releasever && \
    dnf clean all && \
    dnf -y update sqlite-libs libxml2 && \
    dnf clean all

COPY --from=build /app/schedule-runner ./schedule-runner
COPY --link lang ./lang

ENTRYPOINT ["./schedule-runner"]
