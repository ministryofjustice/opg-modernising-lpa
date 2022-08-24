go-test:
	find . -name go.mod -execdir go test ./... -race -covermode=atomic -coverprofile=coverage.out \;

build-up-app:
	docker compose up -d --build --remove-orphans app

build-up-app-testing:
	docker compose -f ./docker-compose.yml \
	-f ./docker-compose.testing.yml \
 	up -d --build app sign-in-mock cypress

run-cypress-dc:
	docker compose -f ./docker-compose.yml \
	-f ./docker-compose.testing.yml \
	run --rm cypress
