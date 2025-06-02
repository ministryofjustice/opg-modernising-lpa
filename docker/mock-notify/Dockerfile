FROM golang:1.24.3-alpine@sha256:b4f875e650466fa0fe62c6fd3f02517a392123eea85f1d7e69d85f780e4db1c1 AS build-env

WORKDIR /app

COPY --link go.mod go.sum ./
RUN go mod download

COPY --link cmd/mock-notify ./cmd/mock-notify

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -a -installsuffix cgo -o /go/bin/mock-notify ./cmd/mock-notify

FROM scratch AS production

WORKDIR /go/bin

COPY --from=build-env /go/bin/mock-notify mock-notify

CMD [ "./mock-notify" ]
