FROM golang:1.21.4-alpine as base

ARG ARCH=amd64

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/event-logger ./cmd/event-logger

RUN CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} go build -o /go/bin/event-logger ./cmd/event-logger

FROM scratch as production

COPY --from=base /go/bin/event-logger event-logger

ENTRYPOINT ["./event-logger"]
