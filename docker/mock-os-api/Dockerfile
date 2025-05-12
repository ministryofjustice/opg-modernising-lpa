FROM golang:1.24.3-alpine@sha256:ef18ee7117463ac1055f5a370ed18b8750f01589f13ea0b48642f5792b234044 AS build-env

WORKDIR /app

COPY --link go.mod go.sum ./
RUN go mod download

COPY --link cmd/mock-os-api ./cmd/mock-os-api

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -a -installsuffix cgo -o /go/bin/mock-os-api ./cmd/mock-os-api

FROM scratch AS production

WORKDIR /go/bin

COPY --from=build-env /go/bin/mock-os-api mock-os-api
COPY --link docker/mock-os-api/data data

CMD [ "./mock-os-api" ]
