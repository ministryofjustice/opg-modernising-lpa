FROM golang:1.26.0-alpine@sha256:d4c4845f5d60c6a974c6000ce58ae079328d03ab7f721a0734277e69905473e5 AS build-env

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
