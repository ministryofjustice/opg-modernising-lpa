FROM golang:1.20 as build

WORKDIR /app

COPY ./cmd/event_received/main.go .
COPY go.mod go.sum ./

RUN go mod download

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -tags lambda.norpc -o main main.go

FROM public.ecr.aws/lambda/provided:al2

COPY --from=build /app/main ./main

ENTRYPOINT ["./main"]
