FROM golang:1.23.1-alpine AS build

WORKDIR /app

COPY --link go.mod go.sum ./
RUN go mod download

COPY --link cmd/schedule-runner ./cmd/schedule-runner
COPY --link internal ./internal

RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=0 go build -o schedule-runner ./cmd/schedule-runner

FROM public.ecr.aws/lambda/provided:al2023 AS dev

WORKDIR /app

COPY --from=build /app/schedule-runner /var/task/schedule-runner
COPY --link lang /var/task/lang
COPY --link docker/aws-lambda-rie ./aws-lambda-rie

ENTRYPOINT ["./schedule-runner"]

FROM public.ecr.aws/lambda/provided:al2023 AS production

WORKDIR /app
COPY --link  docker/install_lambda_insights.sh /app/

RUN chmod +x /app/install_lambda_insights.sh \
  && /app/install_lambda_insights.sh ${TARGETPLATFORM}

COPY --from=build /app/schedule-runner ./schedule-runner
COPY --link lang ./lang

ENTRYPOINT ["./schedule-runner"]
