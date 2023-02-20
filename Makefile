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

build-up-app-testing: ##@build Builds the app with testing config
	docker compose -f ./docker-compose.yml \
	-f ./docker-compose.testing.yml \
 	up -d --build app sign-in-mock pay-mock cypress

run-cypress: ##@testing Runs cypress e2e tests. To run a specific spec file pass in spec e.g. make run-cypress spec=start
ifdef spec
	yarn run cypress:run --spec "cypress/e2e/$(spec).cy.js"
else
	yarn run cypress:run --spec "cypress/e2e/**.cy.js"
endif

run-cypress-headed: ##@testing Runs cypress e2e tests in a browser. To run a specific spec file pass in spec e.g. make run-cypress spec=start
ifdef spec
	yarn run cypress:run --spec "cypress/e2e/$(spec).cy.js" --headed --no-exit
else
	yarn run cypress:run --spec "cypress/e2e/**.cy.js" --headed --no-exit
endif

update-secrets-baseline: ##@security Updates detect-secrets baseline file for false possible and dummy secrets added to version control (requires yelp/detect-secrets local installation)
	$(info ${YELLOW}Ensure any newly added leaks in the baseline are false positives or dummy secrets before committing an updated baseline) @echo "\n"  ${WHITE}
	detect-secrets scan --baseline .secrets.baseline

audit-secrets: ##@security Interactive CLI tool for marking discovered as in/valid (requires yelp/detect-secrets local installation)
	detect-secrets audit .secrets.baseline

run-structurizr:
	docker pull structurizr/lite
	docker run -it --rm -p 8080:8080 -v $(PWD)/docs/architecture/dsl:/usr/local/structurizr structurizr/lite

run-structurizr-mlpa:
	docker pull structurizr/lite
	docker run -it --rm -p 8080:8080 -v $(PWD)/docs/architecture/dsl-mlpa:/usr/local/structurizr structurizr/lite
