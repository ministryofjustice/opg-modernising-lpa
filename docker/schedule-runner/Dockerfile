FROM golang:1.26.1-alpine@sha256:2389ebfa5b7f43eeafbd6be0c3700cc46690ef842ad962f6c5bd6be49ed82039 AS build

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
