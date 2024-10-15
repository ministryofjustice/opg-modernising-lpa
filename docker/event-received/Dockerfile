FROM golang:1.23.2-alpine AS build

WORKDIR /app

COPY --link go.mod go.sum ./
RUN go mod download

COPY --link cmd/event-received ./cmd/event-received
COPY --link internal ./internal

RUN GOOS=linux GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=0 go build -tags lambda.norpc -o event-received ./cmd/event-received

FROM public.ecr.aws/lambda/provided:al2023.2024.10.14.12 AS dev

WORKDIR /app

COPY --from=build /app/event-received /var/task/event-received
COPY --link lang ./lang
COPY --link docker/event-received/aws-lambda-rie ./aws-lambda-rie

ENTRYPOINT ["./event-received"]

FROM public.ecr.aws/lambda/provided:al2023.2024.10.14.12 AS production

WORKDIR /app
COPY --link  docker/event-received/install_lambda_insights.sh /app/

RUN chmod +x /app/install_lambda_insights.sh \
  && /app/install_lambda_insights.sh "${TARGETPLATFORM}"

COPY --from=build /app/event-received ./event-received
COPY --link lang ./lang

# RUN addgroup -S app && \
#   adduser -S -g app app && \
#   chown -R app:app event-received
# USER app
ENTRYPOINT ["./event-received"]
