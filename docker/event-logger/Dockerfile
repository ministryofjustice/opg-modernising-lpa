FROM golang:1.26.3-alpine@sha256:91eda9776261207ea25fd06b5b7fed8d397dd2c0a283e77f2ab6e91bfa71079d AS base

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
