# Uploads S3 Bucket Module

This module creates an S3 bucket for storing uploads, triggers for virus scanning, S3 object replication to the case management application account.

<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.5.2 |
| <a name="requirement_aws"></a> [aws](#requirement\_aws) | ~> 5.42.0 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_aws.region"></a> [aws.region](#provider\_aws.region) | ~> 5.42.0 |

## Modules

| Name | Source | Version |
|------|--------|---------|
| <a name="module_s3_create_batch_replication_jobs"></a> [s3\_create\_batch\_replication\_jobs](#module\_s3\_create\_batch\_replication\_jobs) | ../lambda | n/a |

## Resources

| Name | Type |
|------|------|
| [aws_cloudwatch_metric_alarm.replication-failed](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/cloudwatch_metric_alarm) | resource |
| [aws_iam_policy.replication](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_policy) | resource |
| [aws_iam_role.scheduler_role](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role) | resource |
| [aws_iam_role_policy.s3_create_batch_replication_jobs](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role_policy) | resource |
| [aws_iam_role_policy.scheduler_invoke_lambda](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role_policy) | resource |
| [aws_iam_role_policy_attachment.cloudwatch_lambda_insights](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role_policy_attachment) | resource |
| [aws_iam_role_policy_attachment.replication](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role_policy_attachment) | resource |
| [aws_lambda_permission.av_scan](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/lambda_permission) | resource |
| [aws_lambda_permission.object_tagging](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/lambda_permission) | resource |
| [aws_s3_bucket.bucket](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/s3_bucket) | resource |
| [aws_s3_bucket_lifecycle_configuration.main](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/s3_bucket_lifecycle_configuration) | resource |
| [aws_s3_bucket_logging.bucket](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/s3_bucket_logging) | resource |
| [aws_s3_bucket_notification.bucket_notification](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/s3_bucket_notification) | resource |
| [aws_s3_bucket_ownership_controls.bucket_object_ownership](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/s3_bucket_ownership_controls) | resource |
| [aws_s3_bucket_policy.bucket](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/s3_bucket_policy) | resource |
| [aws_s3_bucket_public_access_block.public_access_policy](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/s3_bucket_public_access_block) | resource |
| [aws_s3_bucket_replication_configuration.replication](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/s3_bucket_replication_configuration) | resource |
| [aws_s3_bucket_server_side_encryption_configuration.bucket_encryption_configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/s3_bucket_server_side_encryption_configuration) | resource |
| [aws_s3_bucket_versioning.bucket_versioning](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/s3_bucket_versioning) | resource |
| [aws_scheduler_schedule.invoke_lambda_every_15_minutes](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/scheduler_schedule) | resource |
| [aws_ssm_parameter.s3_batch_configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/ssm_parameter) | resource |
| [aws_caller_identity.current](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/caller_identity) | data source |
| [aws_default_tags.current](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/default_tags) | data source |
| [aws_iam_policy_document.bucket](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_iam_policy_document.replication](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_iam_policy_document.s3_create_batch_replication_jobs](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_iam_policy_document.scheduler_assume_role](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_iam_policy_document.scheduler_invoke_lambda](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_iam_role.replication](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_role) | data source |
| [aws_kms_alias.cloudwatch_application_logs_encryption](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/kms_alias) | data source |
| [aws_kms_alias.reduced_fees_uploads_s3_encryption](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/kms_alias) | data source |
| [aws_region.current](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/region) | data source |
| [aws_s3_bucket.access_log](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/s3_bucket) | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_bucket_name"></a> [bucket\_name](#input\_bucket\_name) | Name of the bucket. do not use dots (.) except for buckets that are used only for static website hosting. | `string` | n/a | yes |
| <a name="input_create_s3_batch_replication_jobs_lambda_iam_role"></a> [create\_s3\_batch\_replication\_jobs\_lambda\_iam\_role](#input\_create\_s3\_batch\_replication\_jobs\_lambda\_iam\_role) | Lambda IAM role | `any` | n/a | yes |
| <a name="input_events_received_lambda_function"></a> [events\_received\_lambda\_function](#input\_events\_received\_lambda\_function) | Lambda function ARN for events received | `any` | n/a | yes |
| <a name="input_force_destroy"></a> [force\_destroy](#input\_force\_destroy) | A boolean that indicates all objects should be deleted from the bucket so that the bucket can be destroyed without error. These objects are not recoverable. | `bool` | `false` | no |
| <a name="input_s3_antivirus_lambda_function"></a> [s3\_antivirus\_lambda\_function](#input\_s3\_antivirus\_lambda\_function) | Lambda function ARN for events received | `any` | n/a | yes |
| <a name="input_s3_replication"></a> [s3\_replication](#input\_s3\_replication) | s3\_replication = {<br>      enabled                                   = "Enable S3 object replication"<br>      destination\_bucket\_arn                    = "ARN of the destination bucket"<br>      destination\_encryption\_key\_arn            = "ARN of the destination encryption key"<br>      destination\_account\_id                    = "Account ID of the destination bucket"<br>      lambda\_function\_image\_ecr\_arn             = "ARN of the lambda function to be invoked on a schedule to create replication jobs"<br>      lambda\_function\_image\_ecr\_url             = "URL of the lambda function to be invoked on a schedule to create replication jobs"<br>      lambda\_function\_image\_tag                 = "Tag of the lambda function to be invoked on a schedule to create replication jobs"<br>      enable\_s3\_batch\_job\_replication\_scheduler = "Enable scheduler to create replication jobs"<br>    } | <pre>object({<br>    enabled                                   = bool<br>    destination_bucket_arn                    = string<br>    destination_encryption_key_arn            = string<br>    destination_account_id                    = string<br>    lambda_function_image_ecr_arn             = string<br>    lambda_function_image_ecr_url             = string<br>    lambda_function_image_tag                 = string<br>    enable_s3_batch_job_replication_scheduler = bool<br>  })</pre> | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_bucket"></a> [bucket](#output\_bucket) | S3 uploads bucket. |
<!-- END_TF_DOCS -->
