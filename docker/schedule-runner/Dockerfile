FROM golang:1.23.1-alpine AS build

WORKDIR /app

COPY --link go.mod go.sum ./
RUN go mod download

COPY --link cmd/schedule-runner ./cmd/schedule-runner
COPY --link internal ./internal

RUN GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o schedule-runner ./cmd/schedule-runner

FROM public.ecr.aws/lambda/provided:al2023 AS dev

WORKDIR /app

COPY --from=build /app/schedule-runner /var/task/schedule-runner
COPY --link lang ./lang
COPY --link docker/schedule-runner/aws-lambda-rie ./aws-lambda-rie

ENTRYPOINT ["./schedule-runner"]

FROM public.ecr.aws/lambda/provided:al2 AS production

WORKDIR /app

COPY --from=build /app/schedule-runner ./schedule-runner
COPY --link lang ./lang

ENTRYPOINT ["./schedule-runner"]
