FROM golang:1.24.0-alpine AS build

WORKDIR /app

COPY --link go.mod go.sum ./
RUN go mod download

COPY --link cmd/scheduled-task-adder/main.go ./cmd/main.go
COPY --link internal ./internal

RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=0 go build -o scheduled-task-adder ./cmd/main.go

FROM public.ecr.aws/lambda/provided:al2.2025.02.18.01

WORKDIR /app

COPY --from=build /app/scheduled-task-adder /var/task/scheduled-task-adder

ENTRYPOINT ["./scheduled-task-adder"]
