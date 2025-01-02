FROM golang:1.23.4-alpine AS build-env

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
