FROM golang:1.23.3-alpine AS build-env

WORKDIR /app

RUN apk --no-cache add openssl=3.3.2-r1

COPY --link go.mod go.sum ./
RUN go mod download

COPY --link cmd/mock-notify ./cmd/mock-notify

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -a -installsuffix cgo -o /go/bin/mock-notify ./cmd/mock-notify 

FROM scratch AS production

WORKDIR /go/bin

COPY --from=build-env /go/bin/mock-notify mock-notify

CMD [ "./mock-notify" ]
