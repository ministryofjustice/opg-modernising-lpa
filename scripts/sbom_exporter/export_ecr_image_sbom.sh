#!/bin/bash
set -e

# Check if both arguments are provided
if [ -z "$1" ] || [ -z "$2" ] || [ -z "$3" ]; then
  echo "Usage: $0 <service_name> <image_tag> <filter_criteria_file>"
  exit 1
fi

# Use the provided arguments
SERVICE_NAME=$1
IMAGE_TAG=$2
FILTER_CRITERIA_FILE=$3
ACCOUNT_ID=311462405659

echo "Using image tag: $IMAGE_TAG"
echo "Using filter criteria file: $FILTER_CRITERIA_FILE"

# Update the filter_criteria.json with the new IMAGE_TAG
jq --arg tag "$IMAGE_TAG" '.ecrImageTags = [{"comparison": "EQUALS", "value": $tag}]' $FILTER_CRITERIA_FILE > tmp.$$.json

# Create a SBOM export
REQUEST=$(aws inspector2 create-sbom-export \
    --report-format SPDX_2_3 \
    --resource-filter-criteria file://tmp.$$.json \
    --s3-destination bucketName=opg-aws-inspector-sbom,keyPrefix=$SERVICE_NAME/$IMAGE_TAG,kmsKeyArn=arn:aws:kms:eu-west-1:311462405659:key/mrk-1899eeb57e6045d1a85310e1edda47c9)

rm tmp.$$.json

REPORT_ID=$(echo $REQUEST | jq -r '.reportId')

echo "SBOM export request id: $REPORT_ID"

# Wait for export to complete
while true; do
    RESPONSE=$(aws inspector2 get-sbom-export --report-id $REPORT_ID)
    STATUS=$(echo $RESPONSE | jq -r '.status')

    if [ "$STATUS" != "IN_PROGRESS" ]; then
        echo "Final status: $STATUS"
        mkdir -p exports/$SERVICE_NAME/$IMAGE_TAG
        echo "downloading SBOMs from S3..."
        aws s3 cp s3://opg-aws-inspector-sbom/$SERVICE_NAME/$IMAGE_TAG/SPDX_2_3_outputs_$REPORT_ID/account=$ACCOUNT_ID/resource=AWS_ECR_CONTAINER_IMAGE/ ./exports/$SERVICE_NAME/$IMAGE_TAG --recursive
        echo "replacing : with - ..."
        for f in exports/$SERVICE_NAME/$IMAGE_TAG/*.json; do mv -- "$f" "$(echo "$f" | tr ':' '-')"; done
        break
    fi
    echo "Status is $STATUS. Retrying in 10 seconds..."
    sleep 10
done