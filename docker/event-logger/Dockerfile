FROM golang:1.24.3-alpine@sha256:ef18ee7117463ac1055f5a370ed18b8750f01589f13ea0b48642f5792b234044 AS base

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
