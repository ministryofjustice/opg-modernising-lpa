#!/usr/bin/env bash

set -e

echo 'generating key pair'
openssl genpkey -algorithm RSA -out /tmp/private_key.pem -pkeyopt rsa_keygen_bits:2048
openssl rsa -pubout -in /tmp/private_key.pem -out /tmp/public_key.pem

echo 'setting secrets'
awslocal secretsmanager create-secret --name "private-jwt-key-base64" --secret-string "$(base64 /tmp/private_key.pem)"
awslocal secretsmanager create-secret --name "gov-uk-onelogin-identity-public-key" --secret-string "LS0tLS1CRUdJTiBQVUJMSUMgS0VZLS0tLS0KTUZrd0V3WUhLb1pJemowQ0FRWUlLb1pJemowREFRY0RRZ0FFSlEyVmtpZWtzNW9rSTIxY1Jma0FhOXVxN0t4TQo2bTJqWllCeHBybFVXQlpDRWZ4cTI3cFV0Qzd5aXplVlRiZUVqUnlJaStYalhPQjFBbDhPbHFtaXJnPT0KLS0tLS1FTkQgUFVCTElDIEtFWS0tLS0tCg=="
awslocal secretsmanager create-secret --name "cookie-session-keys" --secret-string "[\"$(head -c32 /dev/random | base64)\"]"
awslocal secretsmanager create-secret --name "gov-uk-pay-api-key" --secret-string "totally-fake-key"
awslocal secretsmanager create-secret --name "os-postcode-lookup-api-key" --secret-string "another-fake-key"
awslocal secretsmanager create-secret --name "yoti-private-key" --secret-string "bm90aGluZwo="
awslocal secretsmanager create-secret --name "gov-uk-notify-api-key" --secret-string "extremely_fake-a-b-c-d-e-f-g-h-i-j"

echo 'creating tables'
awslocal dynamodb create-table --table-name lpas --attribute-definitions AttributeName=PK,AttributeType=S AttributeName=SK,AttributeType=S AttributeName=UID,AttributeType=S --key-schema AttributeName=PK,KeyType=HASH AttributeName=SK,KeyType=RANGE --provisioned-throughput ReadCapacityUnits=1000,WriteCapacityUnits=1000 --global-secondary-indexes file://dynamodb-lpa-gsi-schema.json
awslocal dynamodb create-table --table-name reduced-fees --attribute-definitions AttributeName=PK,AttributeType=S AttributeName=SK,AttributeType=S --key-schema AttributeName=PK,KeyType=HASH AttributeName=SK,KeyType=RANGE --provisioned-throughput ReadCapacityUnits=1000,WriteCapacityUnits=1000

echo 'creating bucket'
awslocal s3api create-bucket --bucket evidence --create-bucket-configuration LocationConstraint=eu-west-1
