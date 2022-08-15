go-test:
	find . -name go.mod -execdir go test ./... -race -covermode=atomic -coverprofile=coverage.out \;

build-up-app:
	docker compose up -d --build app
