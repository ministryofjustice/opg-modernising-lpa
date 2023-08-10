import os
import logging
import uuid
import json
import boto3
from aws_xray_sdk.core import patch_all, xray_recorder

logger = logging.getLogger()
logger.setLevel(logging.INFO)
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
    response = create_s3_batch_replication_job(client, variables)
    print(f"job ID: {response['JobId']}")
    logger.info(response)

def create_client():
    client = boto3.client(
        's3control',
        region_name='eu-west-1'
        )
    return client

def set_variables():
    environment = os.getenv('ENVIRONMENT')
    ssm_client = boto3.client('ssm')
    parameter = ssm_client.get_parameter(
    Name=f'/modernising-lpa/s3-batch-configuration/{environment}/s3_batch_configuration',
    )
    json_object = json.loads(parameter['Parameter']['Value'])
    json_object.update({"environment": environment})
    return json_object


def create_s3_batch_replication_job(client, variables):
    request_token = str(uuid.uuid4())
    response = client.create_job(
    AccountId=variables['aws_account_id'],
    ConfirmationRequired=False,
    Operation={
        'S3ReplicateObject': {}
    },
    Report={
        'Bucket': variables['report_and_manifests_bucket'],
        'Format': 'Report_CSV_20180820',
        'Enabled': True,
        'ReportScope': 'AllTasks'
    },
    ClientRequestToken=request_token,

    Description=f'S3 replication {variables["environment"]} - python',
    Priority=10,
    RoleArn=variables['role_arn'],
    ManifestGenerator={
        'S3JobManifestGenerator': {
            'ExpectedBucketOwner': variables['aws_account_id'],
            'SourceBucket': variables['source_bucket'],
            'ManifestOutputLocation': {
                'ExpectedManifestBucketOwner': variables['aws_account_id'],
                'Bucket': variables['report_and_manifests_bucket'],
                'ManifestEncryption': {
                    'SSES3': {}
                },
                'ManifestFormat': 'S3InventoryReport_CSV_20211130'
            },
            'Filter': {
                'EligibleForReplication': True,
                'ObjectReplicationStatuses': [
                    'FAILED',
                    'NONE'
                ]
            },
            'EnableManifestOutput': True
            }
        }
    )
    return response

if __name__ == '__main__':
    main()
