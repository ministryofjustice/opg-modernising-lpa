FROM golang:1.26.2-alpine@sha256:f85330846cde1e57ca9ec309382da3b8e6ae3ab943d2739500e08c86393a21b1 AS build

WORKDIR /app

COPY --link go.mod go.sum ./
RUN go mod download

COPY --link cmd/create-s3-replication-job ./cmd/create-s3-replication-job
COPY --link internal ./internal

RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=0 go build -o create-s3-replication-job ./cmd/create-s3-replication-job

FROM scratch AS production

WORKDIR /app

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /app/create-s3-replication-job ./create-s3-replication-job
COPY --link lang ./lang

ENTRYPOINT ["./create-s3-replication-job"]
