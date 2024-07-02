#! /usr/bin/env bash
# Bash script to pull S3 Antivirus Scan Zip Packages for Lambda
#

set -e

key="/opg-s3-antivirus/zip-version-main"
value=$(aws-vault exec management-operator -- aws ssm get-parameter --name "$key" --query 'Parameter.Value' --output text 2>/dev/null || true)
echo "Using $key: $value"

echo "Pulling antivirus lambda zip and layer version: $value"
wget -q -O ./region/modules/s3_antivirus/lambda_layer.zip https://github.com/ministryofjustice/opg-s3-antivirus/releases/download/"$value"/lambda_layer-amd64.zip
wget -q -O ./region/modules/s3_antivirus/lambda_layer.zip.sha256sum https://github.com/ministryofjustice/opg-s3-antivirus/releases/download/"$value"/lambda_layer-amd64.zip.sha256sum
(cd ./region/modules/s3_antivirus/ && sha256sum -c "lambda_layer.zip.sha256sum")
wget -q -O ./region/modules/s3_antivirus/myFunction.zip https://github.com/ministryofjustice/opg-s3-antivirus/releases/download/"$value"/myFunction-amd64.zip
wget -q -O ./region/modules/s3_antivirus/myFunction.zip.sha256sum https://github.com/ministryofjustice/opg-s3-antivirus/releases/download/"$value"/myFunction-amd64.zip.sha256sum
(cd ./region/modules/s3_antivirus/ && sha256sum -c "myFunction.zip.sha256sum")
