FROM golang:1.21.3-alpine as build

RUN curl -O https://lambda-insights-extension.s3-ap-northeast-1.amazonaws.com/amazon_linux/lambda-insights-extension.rpm && \
    rpm -U lambda-insights-extension.rpm && \
    rm -f lambda-insights-extension.rpm

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/event-received ./cmd/event-received
COPY internal ./internal

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -tags lambda.norpc -o event-received ./cmd/event-received

FROM public.ecr.aws/lambda/provided:al2 AS dev

COPY --from=build /app/event-received ./event-received
COPY lang ./lang
COPY docker/event-received/aws-lambda-rie ./aws-lambda-rie

ENTRYPOINT ["./event-received"]

FROM public.ecr.aws/lambda/provided:al2 AS production

COPY --from=build /app/event-received ./event-received
COPY lang ./lang

ENTRYPOINT ["./event-received"]
