FROM golang:1.24.6-alpine@sha256:c8c5f95d64aa79b6547f3b626eb84b16a7ce18a139e3e9ca19a8c078b85ba80d AS base

ARG ARCH=amd64

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/event-logger ./cmd/event-logger

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /go/bin/event-logger ./cmd/event-logger

FROM scratch AS production

WORKDIR /app

COPY --from=base /go/bin/event-logger event-logger

ENTRYPOINT ["./event-logger"]
