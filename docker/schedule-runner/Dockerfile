FROM golang:1.27.1-alpine@sha256:8a5910f31396cd4d89662f56c68b3ae31d374308270a1c3bd96672ee5ed43414 AS build

WORKDIR /app

COPY --link go.mod go.sum ./
RUN go mod download

COPY --link cmd/schedule-runner ./cmd/schedule-runner
COPY --link internal ./internal

RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=0 go build -ldflags="-X main.Tag=${TAG}" -o schedule-runner ./cmd/schedule-runner

FROM scratch AS production

WORKDIR /app

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /app/schedule-runner ./schedule-runner
COPY --link lang ./lang

ENTRYPOINT ["./schedule-runner"]
