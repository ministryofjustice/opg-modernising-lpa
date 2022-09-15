set -e

openssl genpkey -algorithm RSA -out private_key.pem -pkeyopt rsa_keygen_bits:2048
openssl rsa -pubout -in private_key.pem -out public_key.pem

awslocal secretsmanager create-secret --name "private-jwt-key-base64" --secret-string "$(base64 private_key.pem)"
awslocal secretsmanager create-secret --name "cookie-session-keys" --secret-string "[\"$(head -c32 /dev/random | base64)\"]"
awslocal secretsmanager create-secret --name "gov-uk-pay-api-key" --secret-string "totally-fake-key"

awslocal dynamodb create-table --table-name lpas --attribute-definitions AttributeName=Id,AttributeType=S --key-schema AttributeName=Id,KeyType=HASH --provisioned-throughput ReadCapacityUnits=1000,WriteCapacityUnits=1000

rm private_key.pem public_key.pem
