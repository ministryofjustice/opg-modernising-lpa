go-test:
	find . -name go.mod -execdir go test ./... -race -covermode=atomic -coverprofile=coverage.out \;

run-cypress-dc:
	docker compose -f /Users/andrewpearce/workspace/opg-modernising-lpa/docker-compose.yml \
	-f /Users/andrewpearce/workspace/opg-modernising-lpa/docker-compose.override.yml \
	-f /Users/andrewpearce/workspace/opg-modernising-lpa/docker-compose.testing.yml \
	up cypress

up-dc:
	docker compose up -d app
