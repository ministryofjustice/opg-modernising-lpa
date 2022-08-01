FROM golang:1.18 as build-env

WORKDIR /app

COPY go.mod .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/mlpab

FROM alpine:3.13

COPY --from=build-env /go/bin/mlpab /go/bin/mlpab
ENTRYPOINT ["./go/bin/mlpab"]
