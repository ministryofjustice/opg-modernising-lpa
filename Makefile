#COLORS
GREEN  := $(shell tput -Txterm setaf 2)
WHITE  := $(shell tput -Txterm setaf 7)
YELLOW := $(shell tput -Txterm setaf 3)
RESET  := $(shell tput -Txterm sgr0)

ECR_LOGIN ?= @aws-vault exec management -- aws ecr get-login-password --region eu-west-1 | docker login --username AWS --password-stdin 311462405659.dkr.ecr.eu-west-1.amazonaws.com

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

go-test: ##@testing Runs go unit test suite
	go test -short ./... -race -covermode=atomic -coverprofile=coverage.out

go-integration-test: ##@testing Runs go integration test suite
	go test -run TestIntegration ./internal/dynamo

go-generate: ##@testing Runs go generate for mocks and enums
	git ls-files | grep '.*/enum_.*\.go' | xargs rm -f
	go generate ./...
	git ls-files | grep '.*/mock_.*_test\.go' | xargs rm -f
	go tool mockery

update-event-schemas: ##@testing Gets the latest event schemas from OPG event catalog that we have tests for
	sh ./scripts/get_event_schemas.sh

coverage: ##@testing Produces coverage report and launches browser line based coverage explorer. To test a specific internal package pass in the package name e.g. make coverage package=page
ifdef package
	$(eval t="/tmp/go-cover.$(package).tmp")
	$(eval path="./internal/$(package)/...")
else
	$(eval t="/tmp/go-cover.tmp")
	$(eval path="./internal/...")
endif
	go test -short -coverprofile=$(t) $(path) && go tool cover -html=$(t) && unlink $(t)

down: ##@build Takes all containers down
	COMPOSE_DOCKER_CLI_BUILD=1 DOCKER_BUILDKIT=1 DOCKER_DEFAULT_PLATFORM=linux/$(shell go env GOARCH) docker compose -f docker/docker-compose.yml -f docker/docker-compose.dev.yml down

up: ##@build Builds and brings the app up
	COMPOSE_DOCKER_CLI_BUILD=1 DOCKER_BUILDKIT=1 DOCKER_DEFAULT_PLATFORM=linux/$(shell go env GOARCH) docker compose -f docker/docker-compose.yml up -d --build --remove-orphans app

up-dev: ##@build Builds the app and brings up via Air hot reload with Delve debugging enabled using amd binaries
	COMPOSE_DOCKER_CLI_BUILD=1 DOCKER_BUILDKIT=1 DOCKER_DEFAULT_PLATFORM=linux/$(shell go env GOARCH) docker compose -f docker/docker-compose.yml -f docker/docker-compose.dev.yml up -d --build --force-recreate --remove-orphans app

pull-latest-mock-onelogin: ## @build logs in to management AWS account and pulls the latest mock-onelogin image (assumes ~/.aws/config contains a profile called management)
	aws-vault exec management -- aws ecr get-login-password --region eu-west-1 | docker login --username AWS --password-stdin 311462405659.dkr.ecr.eu-west-1.amazonaws.com
	docker compose -f docker/docker-compose.yml pull mock-onelogin

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

scan-dynamo: ##@dynamodb dumps all entries in the supplied dynamodb table (defaults to Lpas) e.g. scan-dynamo table=lpas-test
ifdef table
	docker compose -f docker/docker-compose.yml exec localstack awslocal dynamodb --region eu-west-1 scan --table-name $(table)
else
	docker compose -f docker/docker-compose.yml exec localstack awslocal dynamodb --region eu-west-1 scan --table-name Lpas
endif

get-lpa: ##@dynamodb dumps all entries in the Lpas dynamodb table that are related to the LPA id supplied e.g. get-lpa id=abc-123
	docker compose -f docker/docker-compose.yml exec localstack awslocal dynamodb --region eu-west-1 \
		query --table-name Lpas --key-condition-expression 'PK = :pk' --expression-attribute-values '{":pk": {"S": "LPA#$(id)"}}'

get-donor-session-id: ##@dynamodb get donor session id by the LPA id supplied e.g. get-donor-session-id lpaId=abc-123
	docker compose -f docker/docker-compose.yml exec localstack awslocal dynamodb --region eu-west-1 \
		query --table-name Lpas --key-condition-expression 'PK = :pk and begins_with(SK, :sk)' --expression-attribute-values '{":pk": {"S": "LPA#$(lpaId)"}, ":sk": {"S": "DONOR#"}}' | jq -r .Items[0].SK.S | sed 's/DONOR#//g'

get-documents:  ##@dynamodb dumps all documents in the lpas dynamodb table that are related to the LPA id supplied e.g. get-documents lpaId=abc-123
	docker compose -f docker/docker-compose.yml exec localstack awslocal dynamodb --region eu-west-1 \
		query --table-name Lpas --key-condition-expression 'PK = :pk and begins_with(SK, :sk)' --expression-attribute-values '{":pk": {"S": "LPA#$(lpaId)"}, ":sk": {"S": "DOCUMENT#"}}'

get-org-members:  ##@dynamodb dumps all members of an org by orgId supplied e.g. get-org-members orgId=abc-123
	docker compose -f docker/docker-compose.yml exec localstack awslocal dynamodb --region eu-west-1 \
		query --table-name Lpas --key-condition-expression 'PK = :pk and begins_with(SK, :sk)' --expression-attribute-values '{":pk": {"S": "ORGANISATION#$(orgId)"}, ":sk": {"S": "MEMBER#"}}'

delete-all-items: ##@dynamodb deletes and recreates Lpas and Sessions dynamodb table
	docker compose -f docker/docker-compose.yml exec localstack awslocal dynamodb --region eu-west-1 \
		delete-table --table-name Lpas || true

	docker compose -f docker/docker-compose.yml exec localstack awslocal dynamodb create-table \
						 --region eu-west-1 \
						 --table-name Lpas \
						 --attribute-definitions AttributeName=PK,AttributeType=S AttributeName=SK,AttributeType=S AttributeName=LpaUID,AttributeType=S AttributeName=UpdatedAt,AttributeType=S \
						 --key-schema AttributeName=PK,KeyType=HASH AttributeName=SK,KeyType=RANGE \
						 --provisioned-throughput ReadCapacityUnits=1000,WriteCapacityUnits=1000 \
						 --global-secondary-indexes file:///usr/dynamodb-lpa-gsi-schema.json

	docker compose -f docker/docker-compose.yml exec localstack awslocal dynamodb --region eu-west-1 \
		delete-table --table-name Sessions || true

	docker compose -f docker/docker-compose.yml exec localstack awslocal dynamodb create-table \
						  --region eu-west-1 \
						  --table-name Sessions \
						  --attribute-definitions AttributeName=PK,AttributeType=S AttributeName=SK,AttributeType=S \
						  --key-schema AttributeName=PK,KeyType=HASH AttributeName=SK,KeyType=RANGE \
						  --provisioned-throughput ReadCapacityUnits=1000,WriteCapacityUnits=1000

emit-evidence-received: ##@events emits an evidence-received event with the given LpaUID e.g. emit-evidence-received uid=abc-123
	curl -X POST "localhost:9001/emit/opg.poas.sirius/evidence-received" \
		-H "Content-Type: application/json" \
		-d '{"uid": "'${uid}'"}'

emit-reduced-fee-approved: ##@events emits a reduced-fee-approved event with the given LpaUID and fee type e.g. emit-reduced-fee-approved uid=abc-123 type=HalfFee
	curl -X POST "localhost:9001/emit/opg.poas.sirius/reduced-fee-approved" \
		-H "Content-Type: application/json" \
		-d '{"uid": "${uid}", "approvedType": "${type}"}'

emit-reduced-fee-declined: ##@events emits a reduced-fee-declined event with the given LpaUID e.g. emit-reduced-fee-declined uid=abc-123
	curl -X POST "localhost:9001/emit/opg.poas.sirius/reduced-fee-declined" \
		-H "Content-Type: application/json" \
		-d '{"uid": "${uid}"}'

emit-further-info-requested: ##@events emits a further-info-requested event with the given LpaUID e.g. emit-further-info-requested uid=abc-123
	curl -X POST "localhost:9001/emit/opg.poas.sirius/further-info-requested" \
		-H "Content-Type: application/json" \
		-d '{"uid": "${uid}"}'

emit-uid-requested: ##@events emits a uid-requested event with the given detail e.g. emit-uid-requested lpaId=abc sessionId=xyz
	curl -X POST "localhost:9001/emit/opg.poas.makeregister/uid-requested" \
		-H "Content-Type: application/json" \
		-d '"LpaID":"${lpaUid}","DonorSessionID":"${sessionId}","Type":"property-and-affairs","Donor":{"Name":"abc","Dob":"2000-01-01","Postcode":"F1 1FF"}"}'

emit-lpa-updated: ##@events emits an lpa-updated event with the given change type e.g. emit-lpa-updated uid=abc-123 changeType=CANNOT_REGISTER
	curl -X POST "localhost:9001/emit/opg.poas.lpastore/lpa-updated" \
		-H "Content-Type: application/json" \
		-d '{"uid": "${uid}", "changeType":"${changeType}"}'

emit-priority-correspondence-sent: ##@events emits a priority-correspondence-sent event with the given LpaUID e.g. emit-priority-correspondence-sent uid=abc-123
	curl -X POST "localhost:9001/emit/opg.poas.sirius/priority-correspondence-sent" \
		-H "Content-Type: application/json" \
		-d '{"uid": "${uid}", "sentDate":"2024-01-02T12:13:14.000006Z"}'

emit-immaterial-change-confirmed: ##@events emits a immaterial-change-confirmed event with the given LpaUID, actor type an actor UID e.g. emit-immaterial-change-confirmed uid=abc-123 actorType=donor actorUid=def-456
	curl -X POST "localhost:9001/emit/opg.poas.sirius/immaterial-change-confirmed" \
		-H "Content-Type: application/json" \
		-d '{"uid": "${uid}", "actorType": "${actorType}", "actorUID": "${actorUID}"}'

emit-material-change-confirmed: ##@events emits a material-change-confirmed event with the given LpaUID, actor type an actor UID e.g. emit-material-change-confirmed uid=abc-123 actorType=donor actorUid=def-456
	curl -X POST "localhost:9001/emit/opg.poas.sirius/material-change-confirmed" \
		-H "Content-Type: application/json" \
		-d '{"uid": "${uid}", "actorType": "${actorType}", "actorUID": "${actorUID}"}'

emit-certificate-provider-identity-check-failed: ##@events emits a certificate-provider-identity-check-failed event with the given LpaUID e.g. certificate-provider-identity-check-failed uid=abc-123
	curl -X POST "localhost:9001/emit/opg.poas.sirius/certificate-provider-identity-check-failed" \
		-H "Content-Type: application/json" \
		-d '{"uid": "${uid}"}'

emit-certificate-provider-submission-completed: ##@events emits a certificate-provider-submission-completed event with the given LpaUID e.g. certificate-provider-submission-completed uid=abc-123
	curl -X POST "localhost:9001/emit/opg.poas.sirius/certificate-provider-submission-completed" \
		-H "Content-Type: application/json" \
		-d '{"uid": "${uid}"}'

emit-object-tags-added-with-virus: ##@events emits a ObjectTagging:Put event with the given S3 key e.g. emit-object-tags-added-with-virus key=doc/key. Also ensures a tag with GuardDutyMalwareScanStatus exists on an existing object set to THREATS_FOUND
	docker compose -f docker/docker-compose.yml exec localstack awslocal s3api \
		put-object-tagging --bucket evidence --key $(key) --tagging '{"TagSet": [{ "Key": "GuardDutyMalwareScanStatus", "Value": "THREATS_FOUND" }]}'

	docker compose -f docker/docker-compose.yml exec localstack awslocal lambda invoke \
		--endpoint-url=http://localhost:4566 \
		--region eu-west-1 \
		--function-name event-received text \
		--payload '{"Records":[{"eventSource":"aws:s3","eventTime":"2023-10-23T15:58:33.081Z","eventName":"ObjectTagging:Put","s3":{"bucket":{"name":"uploads-opg-modernising-lpa-eu-west-1"},"object":{"key":"$(key)"}}}]}'

emit-object-tags-added-without-virus: ##@events emits a ObjectTagging:Put event with the given S3 key e.g. emit-object-tags-added-with-virus key=doc/key. Also ensures a tag with GuardDutyMalwareScanStatus exists on an existing object set to NO_THREATS_FOUND
	docker compose -f docker/docker-compose.yml exec localstack awslocal s3api \
		put-object-tagging --bucket evidence --key $(key) --tagging '{"TagSet": [{ "Key": "GuardDutyMalwareScanStatus", "Value": "NO_THREATS_FOUND" }]}'

	docker compose -f docker/docker-compose.yml exec localstack awslocal lambda invoke \
		--endpoint-url=http://localhost:4566 \
		--region eu-west-1 \
		--function-name event-received text \
		--payload '{"Records":[{"eventSource":"aws:s3","eventTime":"2023-10-23T15:58:33.081Z","eventName":"ObjectTagging:Put","s3":{"bucket":{"name":"uploads-opg-modernising-lpa-eu-west-1"},"object":{"key":"$(key)"}}}]}'

set-uploads-clean: ##@events calls emit-object-tags-added-without-virus for all documents on a given lpa e.g. set-uploads-clean lpaId=abc
	for k in $$(docker compose -f docker/docker-compose.yml exec localstack awslocal dynamodb --region eu-west-1 query --table-name Lpas --key-condition-expression 'PK = :pk and begins_with(SK, :sk)' --expression-attribute-values '{":pk": {"S": "LPA#$(lpaId)"}, ":sk": {"S": "DOCUMENT#"}}' | jq -c -r '.Items[] | .Key[]'); do \
		key=$$k $(MAKE) emit-object-tags-added-without-virus ; \
		done

set-uploads-infected: ##@events calls emit-object-tags-added-with-virus for all documents on a given lpa e.g. set-uploads-clean lpaId=abc
	for k in $$(docker compose -f docker/docker-compose.yml exec localstack awslocal dynamodb --region eu-west-1 query --table-name Lpas --key-condition-expression 'PK = :pk and begins_with(SK, :sk)' --expression-attribute-values '{":pk": {"S": "LPA#$(lpaId)"}, ":sk": {"S": "DOCUMENT#"}}' | jq -c -r '.Items[] | .Key[]'); do \
		key=$$k $(MAKE) emit-object-tags-added-with-virus ; \
		done

format-event: ##@events prints an event in the correct format for testing lambdas in AWS console e.g. format-event detailType=uid-requested detail='{"uid":"abc"}'
	$(eval BODY := $(shell echo '{"version":"0","id":"63eb7e5f-1f10-4744-bba9-e16d327c3b98","detail-type":"$(detailType)","source":"opg.poas.sirius","account":"653761790766","time":"2023-08-30T13:40:30Z","region":"eu-west-1","resources":[],"detail":$(detail)}' | sed 's/"/\\"/g'))
	@echo '{"Records": [{"messageId": "19dd0b57-b21e-4ac1-bd88-01bbb068cb78", "body": "$(BODY)"}]}'

tail-logs: ##@app tails logs for app mock-notify, mock-onelogin, mock-lpa-store, mock-uid, mock-pay and localstack
	docker compose --ansi=always -f docker/docker-compose.yml -f docker/docker-compose.dev.yml logs app mock-notify mock-onelogin mock-lpa-store mock-uid mock-pay localstack -f | grep -v -E "(sqs.ReceiveMessage|Backing off)"

terraform-update-docs: ##@terraform updates all terraform-docs managed documentation
	terraform-docs --config terraform/environment/.terraform-docs.yml ./terraform/environment
	terraform-docs --config terraform/environment/region/.terraform-docs.yml ./terraform/environment/region
	terraform-docs --config terraform/environment/global/.terraform-docs.yml ./terraform/environment/global
	terraform-docs --config terraform/account/.terraform-docs.yml ./terraform/account
	terraform-docs --config terraform/account/region/.terraform-docs.yml ./terraform/account/region

delete-all-from-lpa-index: ##@opensearch clears all items from the lpa index
	curl -X POST "http://localhost:9200/lpas/_delete_by_query/?conflicts=proceed&pretty" -H 'Content-Type: application/json' -d '{"query": {"match_all": {}}}'

delete-lpa-index: ##@opensearch deletes the lpa index
	curl -XDELETE "http://localhost:9200/lpas"

add-scheduled-tasks: ##@scheduler adds scheduled tasks and required entities to test schedule (defaults to 10) e.g. add-scheduled-tasks count=100
ifdef count
	docker compose -f docker/docker-compose.yml exec localstack awslocal lambda invoke \
		--endpoint-url=http://localhost:4566 \
		--region eu-west-1 \
		--function-name scheduled-task-adder text \
		--payload '{"taskCount":$(count)}'

else
	docker compose -f docker/docker-compose.yml exec localstack awslocal lambda invoke \
		--endpoint-url=http://localhost:4566 \
		--region eu-west-1 \
		--function-name scheduled-task-adder text \
		--payload '{"taskCount":10}'
endif

run-schedule-runner: ##@scheduler invokes the schedule-runner lambda
	docker compose -f docker/docker-compose.yml exec localstack awslocal lambda invoke \
			 --endpoint-url=http://localhost:4566 \
			 --region eu-west-1 \
			 --function-name schedule-runner text

test-schedule-runner: add-scheduled-tasks run-schedule-runner ##@scheduler seeds scheduled tasks and runs the schedule-runner (defaults to 10 seeded tasks) e.g. test-schedule-runner count=100
	docker compose -f docker/docker-compose.yml exec localstack awslocal cloudwatch get-metric-data \
		--endpoint-url=http://localhost:4566 \
		--region eu-west-1 \
		--metric-data-queries file:///usr/schedule-runner-metrics-query.json \
		--start-time "$(shell date -v-1H -u +"%Y-%m-%dT%H:%M:%SZ")" \
		--end-time "$(shell date -v+1M -u +"%Y-%m-%dT%H:%M:%SZ")"
