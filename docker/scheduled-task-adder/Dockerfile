FROM golang:1.24.1-alpine@sha256:43c094ad24b6ac0546c62193baeb3e6e49ce14d3250845d166c77c25f64b0386 AS build

WORKDIR /app

COPY --link go.mod go.sum ./
RUN go mod download

COPY --link cmd/scheduled-task-adder/main.go ./cmd/main.go
COPY --link internal ./internal

RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=0 go build -o scheduled-task-adder ./cmd/main.go

FROM public.ecr.aws/lambda/provided:al2023.2025.02.18.01@sha256:edcd0b012f44ce1e87bf865addb662d4be0d62b3a9d8ee72e3c4284a733343b6

WORKDIR /app

COPY --from=build /app/scheduled-task-adder /var/task/scheduled-task-adder

ENTRYPOINT ["./scheduled-task-adder"]
