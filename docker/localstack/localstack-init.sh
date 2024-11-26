#!/usr/bin/env bash
echo 'configuring opensearch'
awslocal opensearch create-domain --region eu-west-1 --domain-name my-domain

echo 'deleting opensearch lpas index'
curl -XDELETE "http://opensearch:9200/lpas"

echo 'generating key pair'
openssl genpkey -algorithm RSA -out /tmp/private_key.pem -pkeyopt rsa_keygen_bits:2048
openssl rsa -pubout -in /tmp/private_key.pem -out /tmp/public_key.pem

echo 'setting secrets'
awslocal secretsmanager create-secret --region eu-west-1 --name "private-jwt-key-base64" --secret-string "$(base64 /tmp/private_key.pem)"
awslocal secretsmanager create-secret --region eu-west-1 --name "cookie-session-keys" --secret-string "[\"$(head -c32 /dev/random | base64)\"]"
awslocal secretsmanager create-secret --region eu-west-1 --name "gov-uk-pay-api-key" --secret-string "totally-fake-key"
awslocal secretsmanager create-secret --region eu-west-1 --name "os-postcode-lookup-api-key" --secret-string "another-fake-key"
awslocal secretsmanager create-secret --region eu-west-1 --name "gov-uk-notify-api-key" --secret-string "extremely_fake-a-b-c-d-e-f-g-h-i-j"
awslocal secretsmanager create-secret --region eu-west-1 --name "lpa-store-jwt-secret-key" --secret-string "more-fake-keys"

echo 'creating tables'
awslocal dynamodb create-table \
         --region eu-west-1 \
         --table-name lpas \
         --attribute-definitions AttributeName=PK,AttributeType=S AttributeName=SK,AttributeType=S AttributeName=LpaUID,AttributeType=S AttributeName=UpdatedAt,AttributeType=S \
         --key-schema AttributeName=PK,KeyType=HASH AttributeName=SK,KeyType=RANGE \
         --provisioned-throughput ReadCapacityUnits=1000,WriteCapacityUnits=1000 \
         --global-secondary-indexes file://dynamodb-lpa-gsi-schema.json

awslocal dynamodb create-table \
         --region eu-west-1 \
         --table-name lpas-test \
         --attribute-definitions AttributeName=PK,AttributeType=S AttributeName=SK,AttributeType=S AttributeName=LpaUID,AttributeType=S AttributeName=UpdatedAt,AttributeType=S \
         --key-schema AttributeName=PK,KeyType=HASH AttributeName=SK,KeyType=RANGE \
         --provisioned-throughput ReadCapacityUnits=1000,WriteCapacityUnits=1000 \
         --global-secondary-indexes file://dynamodb-lpa-gsi-schema.json

echo 'creating bucket'
awslocal s3api create-bucket --bucket evidence --create-bucket-configuration LocationConstraint=eu-west-1

echo 'configuring events'
awslocal sqs create-queue --region eu-west-1 --queue-name event-queue
awslocal sqs create-queue --region eu-west-1 --queue-name event-bus-queue
awslocal events create-event-bus --region eu-west-1 --name default

awslocal events put-rule \
  --region eu-west-1 \
  --name send-events-to-queue-rule \
  --event-bus-name default \
  --event-pattern '{}'

awslocal events put-targets \
  --region eu-west-1 \
  --event-bus-name default \
  --rule send-events-to-queue-rule \
  --targets "Id"="event-queue","Arn"="arn:aws:sqs:eu-west-1:000000000000:event-queue"

awslocal events put-rule \
  --region eu-west-1 \
  --name send-events-to-bus-queue-rule \
  --event-bus-name default \
  --event-pattern '{"source":["opg.poas.makeregister"],"detail-type":["uid-requested"]}'

awslocal events put-targets \
  --region eu-west-1 \
  --event-bus-name default \
  --rule send-events-to-bus-queue-rule \
  --targets "Id"="event-bus-queue","Arn"="arn:aws:sqs:eu-west-1:000000000000:event-bus-queue"

echo 'creating log groups'
awslocal logs create-log-group \
  --region eu-west-1 \
  --log-group-name "local_application_logs"

echo 'creating event-received lambda'
awslocal lambda create-function \
  --environment Variables="{LPAS_TABLE=$LPAS_TABLE,GOVUK_NOTIFY_IS_PRODUCTION=$GOVUK_NOTIFY_IS_PRODUCTION,APP_PUBLIC_URL=$APP_PUBLIC_URL,GOVUK_NOTIFY_BASE_URL=$GOVUK_NOTIFY_BASE_URL,UPLOADS_S3_BUCKET_NAME=$UPLOAD_S3_BUCKET,UID_BASE_URL=$UID_BASE_URL,SEARCH_ENDPOINT=$SEARCH_ENDPOINT,SEARCH_INDEXING_DISABLED=$SEARCH_INDEXING_DISABLED}" \
  --region eu-west-1 \
  --function-name event-received \
  --handler bootstrap \
  --runtime provided.al2023 \
  --role arn:aws:iam::000000000000:role/lambda-role \
  --zip-file fileb:///etc/event-received.zip

awslocal lambda wait function-active-v2 --region eu-west-1 --function-name event-received

echo 'creating schedule-runner lambda'
awslocal lambda create-function \
  --environment Variables="{AWS_BASE_URL=$AWS_BASE_URL,TAG=$TAG,LPAS_TABLE=$LPAS_TABLE,GOVUK_NOTIFY_IS_PRODUCTION=$GOVUK_NOTIFY_IS_PRODUCTION,GOVUK_NOTIFY_BASE_URL=$GOVUK_NOTIFY_BASE_URL,SEARCH_ENDPOINT=$SEARCH_ENDPOINT,SEARCH_INDEXING_DISABLED=$SEARCH_INDEXING_DISABLED,SEARCH_INDEX_NAME=$SEARCH_INDEX_NAME,XRAY_ENABLED=$XRAY_ENABLED,METRICS_ENABLED=$METRICS_ENABLED,LPA_STORE_BASE_URL=$LPA_STORE_BASE_URL,LPA_STORE_SECRET_ARN=$LPA_STORE_SECRET_ARN}" \
  --region eu-west-1 \
  --function-name schedule-runner \
  --handler bootstrap \
  --runtime provided.al2023 \
  --role arn:aws:iam::000000000000:role/lambda-role \
  --zip-file fileb:///etc/schedule-runner.zip \
  --timeout 900 \
  --tracing-config Mode=Active \
  --logging-config LogGroup="local_application_logs"

awslocal lambda wait function-active-v2 --region eu-west-1 --function-name schedule-runner

echo 'creating scheduled-task-adder lambda'
awslocal lambda create-function \
  --environment Variables="{AWS_BASE_URL=$AWS_BASE_URL}" \
  --region eu-west-1 \
  --function-name scheduled-task-adder \
  --handler bootstrap \
  --runtime provided.al2023 \
  --role arn:aws:iam::000000000000:role/lambda-role \
  --zip-file fileb:///etc/scheduled-task-adder.zip \
  --timeout 120

awslocal lambda wait function-active-v2 --region eu-west-1 --function-name scheduled-task-adder

echo 'create and associate scheduler'
awslocal scheduler create-schedule \
  --region eu-west-1 \
  --name schedule-runner-minutely \
  --schedule-expression 'rate(1 minute)' \
  --description "Runs every minute (to aid testing - deployed infra will run less frequently)" \
  --target '{"RoleArn": "arn:aws:iam::000000000000:role/lambda-role", "Arn":"arn:aws:lambda:eu-west-1:000000000000:function:schedule-runner" }' \
  --flexible-time-window '{ "Mode": "OFF"}'


echo 'creating event source mapping'
awslocal lambda create-event-source-mapping \
         --function-name "arn:aws:lambda:eu-west-1:000000000000:function:event-received" \
         --batch-size 1 \
         --event-source-arn "arn:aws:sqs:eu-west-1:000000000000:event-bus-queue"
