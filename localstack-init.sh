set -e

openssl genrsa -out private_key.pem 2048
openssl rsa -in private_key.pem -outform PEM -pubout -out public_key.pem

awslocal secretsmanager create-secret --name "private-jwt-key-base64" --secret-string "$(base64 private_key.pem)"
awslocal secretsmanager create-secret --name "public-jwt-key-base64" --secret-string "$(base64 public_key.pem)"

rm private_key.pem public_key.pem
