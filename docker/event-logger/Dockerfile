FROM golang:1.24.1-alpine@sha256:43c094ad24b6ac0546c62193baeb3e6e49ce14d3250845d166c77c25f64b0386 AS base

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
