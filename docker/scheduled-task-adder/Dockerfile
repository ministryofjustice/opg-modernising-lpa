FROM golang:1.25.1-alpine@sha256:b6ed3fd0452c0e9bcdef5597f29cc1418f61672e9d3a2f55bf02e7222c014abd AS build

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
