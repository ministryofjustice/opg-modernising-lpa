set -e

openssl genpkey -algorithm RSA -out ./private_key.pem -pkeyopt rsa_keygen_bits:2048
openssl rsa -pubout -in ./private_key.pem -out ./public_key.pem

awslocal secretsmanager create-secret --name "default/private-jwt-key-base64" --secret-string "$(base64 private.pem)"
awslocal secretsmanager create-secret --name "default/public-jwt-key-base64" --secret-string "$(base64 public.pem)"

rm private.pem public.pem
