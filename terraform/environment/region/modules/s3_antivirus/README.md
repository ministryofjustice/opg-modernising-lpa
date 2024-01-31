# S3 Antivirus Module

This module deploys a lambda function that scans S3 objects for viruses on put.

## Requirements

| Name                                                                      | Version   |
|---------------------------------------------------------------------------|-----------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.5.2  |
| <a name="requirement_aws"></a> [aws](#requirement\_aws)                   | ~> 5.34.0 |

## Providers

| Name                                                                   | Version   |
|------------------------------------------------------------------------|-----------|
| <a name="provider_aws.region"></a> [aws.region](#provider\_aws.region) | ~> 5.34.0 |

## Modules

No modules.

## Resources

| Name                                                                                                                                                                | Type        |
|---------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------|
| [aws_cloudwatch_log_group.lambda](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/cloudwatch_log_group)                                 | resource    |
| [aws_cloudwatch_metric_alarm.virus_infections](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/cloudwatch_metric_alarm)                 | resource    |
| [aws_iam_role_policy.lambda](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role_policy)                                           | resource    |
| [aws_lambda_alias.lambda_alias](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/lambda_alias)                                           | resource    |
| [aws_lambda_function.lambda_function](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/lambda_function)                                  | resource    |
| [aws_lambda_permission.allow_lambda_execution_operator](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/lambda_permission)              | resource    |
| [aws_lambda_provisioned_concurrency_config.main](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/lambda_provisioned_concurrency_config) | resource    |
| [aws_s3_bucket_metric.virus_infections](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/s3_bucket_metric)                               | resource    |
| [aws_caller_identity.current](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/caller_identity)                                       | data source |
| [aws_default_tags.current](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/default_tags)                                             | data source |
| [aws_iam_policy_document.lambda](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document)                                | data source |
| [aws_kms_alias.cloudwatch_application_logs_encryption](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/kms_alias)                    | data source |
| [aws_kms_alias.uploads_encryption_key](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/kms_alias)                                    | data source |
| [aws_region.current](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/region)                                                         | data source |
| [aws_security_group.lambda_egress](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/security_group)                                   | data source |

## Inputs

| Name                                                                                                                                                 | Description                                                       | Type           | Default | Required |
|------------------------------------------------------------------------------------------------------------------------------------------------------|-------------------------------------------------------------------|----------------|---------|:--------:|
| <a name="input_alarm_sns_topic_arn"></a> [alarm\_sns\_topic\_arn](#input\_alarm\_sns\_topic\_arn)                                                    | ARN of the SNS topic for alarm notifications                      | `string`       | n/a     |   yes    |
| <a name="input_aws_subnet_ids"></a> [aws\_subnet\_ids](#input\_aws\_subnet\_ids)                                                                     | List of Sirius private subnet Ids                                 | `list(string)` | n/a     |   yes    |
| <a name="input_data_store_bucket"></a> [data\_store\_bucket](#input\_data\_store\_bucket)                                                            | Data store bucket to scan for viruses                             | `any`          | n/a     |   yes    |
| <a name="input_definition_bucket"></a> [definition\_bucket](#input\_definition\_bucket)                                                              | Bucket containing virus definitions                               | `any`          | n/a     |   yes    |
| <a name="input_ecr_image_uri"></a> [ecr\_image\_uri](#input\_ecr\_image\_uri)                                                                        | URI of ECR image to use for Lambda                                | `string`       | n/a     |   yes    |
| <a name="input_environment_variables"></a> [environment\_variables](#input\_environment\_variables)                                                  | A map that defines environment variables for the Lambda Function. | `map(string)`  | `{}`    |    no    |
| <a name="input_lambda_task_role"></a> [lambda\_task\_role](#input\_lambda\_task\_role)                                                               | Execution role for Lambda                                         | `any`          | n/a     |   yes    |
| <a name="input_s3_antivirus_provisioned_concurrency"></a> [s3\_antivirus\_provisioned\_concurrency](#input\_s3\_antivirus\_provisioned\_concurrency) | Number of concurrent executions to provision for Lambda           | `number`       | n/a     |   yes    |

## Outputs

| Name                                                                                | Description |
|-------------------------------------------------------------------------------------|-------------|
| <a name="output_lambda_function"></a> [lambda\_function](#output\_lambda\_function) | n/a         |
