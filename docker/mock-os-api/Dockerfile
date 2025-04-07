FROM golang:1.24.2-alpine@sha256:7772cb5322baa875edd74705556d08f0eeca7b9c4b5367754ce3f2f00041ccee AS build-env

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
