#COLORS
GREEN  := $(shell tput -Txterm setaf 2)
WHITE  := $(shell tput -Txterm setaf 7)
YELLOW := $(shell tput -Txterm setaf 3)
RESET  := $(shell tput -Txterm sgr0)

# Add the following 'help' target to your Makefile
# And add help text after each target name starting with '\#\#'
# A category can be added with @category
# This was made possible by https://gist.github.com/prwhite/8168133#gistcomment-1727513
HELP_FUN = \
    %help; \
    while(<>) { push @{$$help{$$2 // 'options'}}, [$$1, $$3] if /^([a-zA-Z0-9\-]+)\s*:.*\#\#(?:@([a-zA-Z\-]+))?\s(.*)$$/ }; \
    print "usage: make [target]\n\n"; \
    for (sort keys %help) { \
    print "${WHITE}$$_:${RESET}\n"; \
    for (@{$$help{$$_}}) { \
    $$sep = " " x (32 - length $$_->[0]); \
    print "  ${YELLOW}$$_->[0]${RESET}$$sep${GREEN}$$_->[1]${RESET}\n"; \
    }; \
    print "\n"; }

help: ##@other Show this help.
	@perl -e '$(HELP_FUN)' $(MAKEFILE_LIST)

go-test: ##@testing Runs full go test suite
	find . -name go.mod -execdir go test ./... -race -covermode=atomic -coverprofile=coverage.out \;

go-generate: ##@testing Runs go generate
	find . -name go.mod -execdir go generate ./... \;

coverage: ##@testing Produces coverage report and launches browser line based coverage explorer. To test a specific internal package pass in the package name e.g. make coverage package=page
ifdef package
	$(eval t="/tmp/go-cover.$(package).tmp")
	$(eval path="./app/internal/$(package)/...")
else
	$(eval t="/tmp/go-cover.tmp")
	$(eval path="./app/...")
endif
	go test -coverprofile=$(t) $(path) && go tool cover -html=$(t) && unlink $(t)

build-up-app: ##@build Builds the app
	docker compose up -d --build --remove-orphans app

build-up-app-dev: ##@build Builds the app and brings up via Air hot reload with Delve debugging enabled using amd binaries
	docker compose -f ./docker-compose.yml -f ./docker-compose.dev.yml up -d --build --force-recreate --remove-orphans app

build-up-app-dev-arm: ##@build Builds the app and brings up via Air hot reload with Delve debugging enabled using arm binaries
	ARCH=arm64 docker compose -f ./docker-compose.yml -f ./docker-compose.dev.yml up -d --build --force-recreate --remove-orphans app

run-cypress: ##@testing Runs cypress e2e tests. To run a specific spec file pass in spec e.g. make run-cypress spec=start
ifdef spec
	yarn run cypress:run --spec "cypress/e2e/$(spec).cy.js"
else
	yarn run cypress:run
endif

run-cypress-headed: ##@testing Runs cypress e2e tests in a browser. To run a specific spec file pass in spec e.g. make run-cypress spec=start
ifdef spec
	yarn run cypress:run --spec "cypress/e2e/$(spec).cy.js" --headed --no-exit
else
	yarn run cypress:run --headed --no-exit
endif

update-secrets-baseline: ##@security Updates detect-secrets baseline file for false possible and dummy secrets added to version control (requires yelp/detect-secrets local installation)
	$(info ${YELLOW}Ensure any newly added leaks in the baseline are false positives or dummy secrets before committing an updated baseline) @echo "\n"  ${WHITE}
	detect-secrets scan --baseline .secrets.baseline

audit-secrets: ##@security Interactive CLI tool for marking discovered as in/valid (requires yelp/detect-secrets local installation)
	detect-secrets audit .secrets.baseline

run-structurizr:
	docker pull structurizr/lite
	docker run -it --rm -p 8081:8080 -v $(PWD)/docs/architecture/dsl/local:/usr/local/structurizr structurizr/lite

run-structurizr-export:
	docker pull structurizr/cli:latest
	docker run --rm -v $(PWD)/docs/architecture/dsl/local:/usr/local/structurizr structurizr/cli \
	export -workspace /usr/local/structurizr/workspace.dsl -format mermaid

scan-lpas: ##@app dumps all entries in the lpas dynamodb table
	docker compose exec localstack awslocal dynamodb scan --table-name lpas

get-lpa:  ##@app dumps all entries in the lpas dynamodb table that are related to the LPA id supplied e.g. get-lpa ID=abc-123
	docker compose exec localstack awslocal dynamodb \
		query --table-name lpas --key-condition-expression 'PK = :pk' --expression-attribute-values '{":pk": {"S": "LPA#$(ID)"}}'

emit-evidence-received: ##@app emits an evidence-received event with the given UID e.g. emit-evidence-received UID=abc-123
	curl "http://localhost:9000/2015-03-31/functions/function/invocations" -d '{"version":"0","id":"63eb7e5f-1f10-4744-bba9-e16d327c3b98","detail-type":"evidence-received","source":"opg.poas.sirius","account":"653761790766","time":"2023-08-30T13:40:30Z","region":"eu-west-1","resources":[],"detail":{"UID":"$(UID)"}}'

emit-fee-approved: ##@app emits a fee-approved event with the given UID e.g. emit-fee-approved UID=abc-123
	curl "http://localhost:9000/2015-03-31/functions/function/invocations" -d '{"version":"0","id":"63eb7e5f-1f10-4744-bba9-e16d327c3b98","detail-type":"fee-approved","source":"opg.poas.sirius","account":"653761790766","time":"2023-08-30T13:40:30Z","region":"eu-west-1","resources":[],"detail":{"UID":"$(UID)"}}'

emit-more-evidence-required: ##@app emits a more-evidence-required event with the given UID e.g. emit-more-evidence-required UID=abc-123
	curl "http://localhost:9000/2015-03-31/functions/function/invocations" -d '{"version":"0","id":"63eb7e5f-1f10-4744-bba9-e16d327c3b98","detail-type":"more-evidence-required","source":"opg.poas.sirius","account":"653761790766","time":"2023-08-30T13:40:30Z","region":"eu-west-1","resources":[],"detail":{"UID":"$(UID)"}}'
