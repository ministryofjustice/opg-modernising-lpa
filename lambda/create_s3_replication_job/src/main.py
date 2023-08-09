import os
import ast
import logging
import boto3
from aws_xray_sdk.core import patch_all, xray_recorder

logger = logging.getLogger()
logger.setLevel(logging.DEBUG)
xray_recorder.begin_segment('reduced_fees_uploads')
patch_all()


def handler(event, context):
    subsegment = xray_recorder.begin_subsegment('batch_replicate_to_sirius')
    subsegment.put_annotation('service', 'reduced_fees_uploads')
    variables = set_variables()
    client = create_client()
    create_s3_batch_replication_job(client, variables)
    xray_recorder.end_subsegment()

def main():
    variables = set_variables()
    client = create_client()
    create_s3_batch_replication_job(client, variables)

def create_client():
    client = boto3.client('s3control')
    return client

def set_variables():
    variables = {
    'aws_account_id':'',
    'report_and_manifests_bucket':'',
    'source_bucket':'',
    'description':'',
    'role_arn':'',
    'aws_region':'1'
    }
    return variables


def create_s3_batch_replication_job(client, variables):
    response = client.create_job(
    AccountId=variables['aws_account_id'],
    ConfirmationRequired=False,
    Operation={
        'S3ReplicateObject': {}
    },
    Report={
        "Bucket": variables['report_and_manifests_bucket'],
        "Format": "Report_CSV_20180820",
        "Enabled": True,
        "ReportScope": "AllTasks"
    },
    ClientRequestToken='string',

    Description=variables['description'],
    Priority=10,
    RoleArn=variables['role_arn'],
    ManifestGenerator={
        "S3JobManifestGenerator": {
            "ExpectedBucketOwner": variables['aws_account_id'],
            "SourceBucket": variables['source_bucket'],
            "ManifestOutputLocation": {
                "ExpectedManifestBucketOwner": variables['aws_account_id'],
                "Bucket": variables['report_and_manifests_bucket'],
                "ManifestEncryption": {
                    "SSES3": {}
                },
                "ManifestFormat": "S3InventoryReport_CSV_20211130"
            },
            "Filter": {
                "EligibleForReplication": True,
                "ObjectReplicationStatuses": [
                    "FAILED",
                    "NONE"
                ]
            },
            "EnableManifestOutput": True
            }
        }
    )
    logger.info(response)

if __name__ == '__main__':
    print(main())
