FROM golang:1.21.4-alpine as build

WORKDIR /app

COPY --link go.mod go.sum ./
RUN go mod download

COPY --link cmd/event-received ./cmd/event-received
COPY --link internal ./internal

RUN GOOS=linux GOARCH=${TARGETPLATFORM} CGO_ENABLED=0 go build -tags lambda.norpc -o event-received ./cmd/event-received

FROM public.ecr.aws/lambda/provided:al2 AS dev

WORKDIR /app

COPY --from=build /app/event-received /var/task/event-received
COPY --link lang ./lang
COPY --link docker/event-received/aws-lambda-rie ./aws-lambda-rie

ENTRYPOINT ["./event-received"]

FROM public.ecr.aws/lambda/provided:al2 AS production

WORKDIR /app

RUN curl -O https://lambda-insights-extension.s3-ap-northeast-1.amazonaws.com/amazon_linux/lambda-insights-extension.rpm && \
    rpm -U lambda-insights-extension.rpm && \
    rm -f lambda-insights-extension.rpm ;

COPY --from=build /app/event-received ./event-received
COPY --link lang ./lang

ENTRYPOINT ["./event-received"]
