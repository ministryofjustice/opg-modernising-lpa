FROM golang:1.27.2-alpine@sha256:f92b6ef800e499660581efdabdf25d9d817a9d124eaf900924f0504e7e27e12d AS build

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
