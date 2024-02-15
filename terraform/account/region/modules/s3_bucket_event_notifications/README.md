# S3 Bucket Event Notifications Module

This module creates a S3 bucket event notifications and event notification filters.

<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.5.2 |
| <a name="requirement_aws"></a> [aws](#requirement\_aws) | ~> 5.36.0 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_aws"></a> [aws](#provider\_aws) | ~> 5.36.0 |

## Modules

No modules.

## Resources

| Name | Type |
|------|------|
| [aws_s3_bucket_notification.bucket_notification](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/s3_bucket_notification) | resource |
| [aws_sns_topic.s3_event_notification](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/sns_topic) | resource |
| [aws_sns_topic_policy.s3_event_notification](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/sns_topic_policy) | resource |
| [aws_iam_policy_document.sns_topic_policy](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_kms_alias.sns](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/kms_alias) | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_s3_bucket_event_types"></a> [s3\_bucket\_event\_types](#input\_s3\_bucket\_event\_types) | The type of event that triggers the notification | `list(string)` | <pre>[<br>  "s3:ObjectRemoved:*",<br>  "s3:ObjectAcl:Put"<br>]</pre> | no |
| <a name="input_s3_bucket_id"></a> [s3\_bucket\_id](#input\_s3\_bucket\_id) | The ID of the S3 bucket to which the notification is attached | `string` | n/a | yes |
| <a name="input_sns_failure_feedback_role_arn"></a> [sns\_failure\_feedback\_role\_arn](#input\_sns\_failure\_feedback\_role\_arn) | The ARN of the IAM role that Amazon SNS can assume when it needs to access your AWS resources to process your failure feedback | `string` | n/a | yes |
| <a name="input_sns_kms_key_alias"></a> [sns\_kms\_key\_alias](#input\_sns\_kms\_key\_alias) | The alias of the KMS key used to encrypt the SNS topic | `string` | n/a | yes |
| <a name="input_sns_success_feedback_role_arn"></a> [sns\_success\_feedback\_role\_arn](#input\_sns\_success\_feedback\_role\_arn) | The ARN of the IAM role that Amazon SNS can assume when it needs to access your AWS resources to process your success feedback | `string` | n/a | yes |

## Outputs

No outputs.
<!-- END_TF_DOCS -->
