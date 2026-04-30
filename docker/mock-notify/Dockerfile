FROM golang:1.26.2-alpine@sha256:f85330846cde1e57ca9ec309382da3b8e6ae3ab943d2739500e08c86393a21b1 AS build-env

WORKDIR /app

COPY --link go.mod go.sum ./
RUN go mod download

COPY --link cmd/mock-notify ./cmd/mock-notify

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -a -installsuffix cgo -o /go/bin/mock-notify ./cmd/mock-notify

FROM scratch AS production

WORKDIR /go/bin

COPY --from=build-env /go/bin/mock-notify mock-notify

CMD [ "./mock-notify" ]
