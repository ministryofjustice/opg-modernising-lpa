FROM golang:1.21 as build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./cmd/event-received/main.go .
COPY ./internal ./internal

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -tags lambda.norpc -o main main.go

FROM public.ecr.aws/lambda/provided:al2 AS dev

COPY --from=build /app/main ./main
COPY ./lang ./lang
COPY ./cmd/event-received/.aws-lambda-rie/aws-lambda-rie ./aws-lambda-rie

ENTRYPOINT ["./main"]

FROM public.ecr.aws/lambda/provided:al2 AS production

COPY --from=build /app/main ./main
COPY ./lang ./lang

ENTRYPOINT ["./main"]
