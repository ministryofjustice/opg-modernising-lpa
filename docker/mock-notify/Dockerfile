FROM golang:1.24.1-alpine@sha256:43c094ad24b6ac0546c62193baeb3e6e49ce14d3250845d166c77c25f64b0386 AS build-env

WORKDIR /app

COPY --link go.mod go.sum ./
RUN go mod download

COPY --link cmd/mock-notify ./cmd/mock-notify

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -a -installsuffix cgo -o /go/bin/mock-notify ./cmd/mock-notify

FROM scratch AS production

WORKDIR /go/bin

COPY --from=build-env /go/bin/mock-notify mock-notify

CMD [ "./mock-notify" ]
