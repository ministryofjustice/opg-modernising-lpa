FROM golang:1.26.2-alpine@sha256:f85330846cde1e57ca9ec309382da3b8e6ae3ab943d2739500e08c86393a21b1 AS build

WORKDIR /app

COPY --link go.mod go.sum ./
RUN go mod download

COPY --link cmd/event-received ./cmd/event-received
COPY --link internal ./internal

RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=0 go build -o event-received ./cmd/event-received

FROM public.ecr.aws/lambda/provided:al2023.2026.02.18.07@sha256:059ed067a8e796d891100888928e32cde75d63d459617260a7901e217a7249d6 AS production

WORKDIR /app
COPY --link  docker/install_lambda_insights.sh /app/

RUN chmod +x /app/install_lambda_insights.sh \
  && /app/install_lambda_insights.sh "${TARGETPLATFORM}"

# Switch DNF to the latest AL2023.10 release (2023.10.20260202) and update SQLite packages
RUN echo "2023.10.20260202" > /etc/dnf/vars/releasever && \
    dnf clean all && \
    dnf -y update sqlite-libs libxml2 && \
    dnf clean all

COPY --from=build /app/event-received ./event-received
COPY --link lang ./lang

ENTRYPOINT ["./event-received"]
