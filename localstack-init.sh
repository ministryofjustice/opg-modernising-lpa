set -e

openssl genpkey -algorithm RSA -out private_key.pem -pkeyopt rsa_keygen_bits:2048
openssl rsa -pubout -in private_key.pem -out public_key.pem

awslocal secretsmanager create-secret --name "private-jwt-key-base64" --secret-string "$(base64 private_key.pem)"
awslocal secretsmanager create-secret --name "cookie-session-keys" --secret-string "[\"$(head -c32 /dev/random | base64)\"]"
awslocal secretsmanager create-secret --name "gov-uk-pay-api-key" --secret-string "totally-fake-key"
awslocal secretsmanager create-secret --name "os-postcode-lookup-api-key" --secret-string "another-fake-key"
awslocal secretsmanager create-secret --name "yoti-private-key" --secret-string "bm90aGluZwo="
awslocal secretsmanager create-secret --name "gov-uk-notify-api-key" --secret-string "extremely_fake-a-b-c-d-e-f-g-h-i-j"

awslocal dynamodb create-table --table-name lpas --attribute-definitions AttributeName=PK,AttributeType=S AttributeName=SK,AttributeType=S --key-schema AttributeName=PK,KeyType=HASH AttributeName=SK,KeyType=RANGE --provisioned-throughput ReadCapacityUnits=1000,WriteCapacityUnits=1000

rm private_key.pem public_key.pem
