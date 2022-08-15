set -e

openssl genpkey -algorithm RSA -out ./private_key.pem -pkeyopt rsa_keygen_bits:2048
openssl rsa -pubout -in ./private_key.pem -out ./public_key.pem

awslocal secretsmanager create-secret --name "private-jwt-key-base64" --secret-string "$(base64 private_key.pem)"
awslocal secretsmanager create-secret --name "public-jwt-key-base64" --secret-string "$(base64 public_key.pem)"

rm private_key.pem public_key.pem
