go-test:
	find . -name go.mod -execdir go test ./... -race -covermode=atomic -coverprofile=coverage.out \;

run-cypress-dc:
	docker compose -f ./docker-compose.yml \
	-f ./docker-compose.testing.yml \
	up cypress

up-dc:
	docker compose up -d app
