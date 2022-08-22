#!/bin/sh

echo "Generating keypair..."

openssl genpkey -algorithm RSA -out ./private_key.pem -pkeyopt rsa_keygen_bits:2048
openssl rsa -pubout -in ./private_key.pem -out ./public_key.pem

chown app:app ./private_key.pem
chown app:app ./public_key.pem

echo "Running proxy..."
exec "$@"
