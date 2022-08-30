set -e

openssl genpkey -algorithm RSA -out private_key.pem -pkeyopt rsa_keygen_bits:2048
openssl rsa -pubout -in private_key.pem -out public_key.pem

awslocal secretsmanager create-secret --name "private-jwt-key-base64" --secret-string "$(base64 private_key.pem)"
awslocal secretsmanager create-secret --name "cookie-session-keys" --secret-string "[\"$(head -c32 /dev/random | base64)\"]"

rm private_key.pem public_key.pem
