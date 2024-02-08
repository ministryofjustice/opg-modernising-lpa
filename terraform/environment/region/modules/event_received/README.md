# Event Received Module

This module creates the resources required to receive and process events from the event bus.

<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.5.2 |
| <a name="requirement_aws"></a> [aws](#requirement\_aws) | ~> 5.35.0 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_aws"></a> [aws](#provider\_aws) | ~> 5.35.0 |
| <a name="provider_aws.region"></a> [aws.region](#provider\_aws.region) | ~> 5.35.0 |

## Modules

| Name | Source | Version |
|------|--------|---------|
| <a name="module_event_received"></a> [event\_received](#module\_event\_received) | ../lambda | n/a |

## Resources

| Name | Type |
|------|------|
| [aws_cloudwatch_event_rule.receive_events_mlpa](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/cloudwatch_event_rule) | resource |
| [aws_cloudwatch_event_rule.receive_events_sirius](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/cloudwatch_event_rule) | resource |
| [aws_cloudwatch_event_target.receive_events_mlpa](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/cloudwatch_event_target) | resource |
| [aws_cloudwatch_event_target.receive_events_sirius](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/cloudwatch_event_target) | resource |
| [aws_iam_role_policy.event_received](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role_policy) | resource |
| [aws_iam_role_policy_attachment.cloudwatch_lambda_insights](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role_policy_attachment) | resource |
| [aws_lambda_permission.allow_cloudwatch_to_call_event_received_mlpa](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/lambda_permission) | resource |
| [aws_lambda_permission.allow_cloudwatch_to_call_event_received_sirius](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/lambda_permission) | resource |
| [aws_caller_identity.current](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/caller_identity) | data source |
| [aws_default_tags.current](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/default_tags) | data source |
| [aws_iam_policy_document.api_access_policy](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_iam_policy_document.event_received](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_kms_alias.cloudwatch_application_logs_encryption](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/kms_alias) | data source |
| [aws_kms_alias.dynamodb_encryption_key](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/kms_alias) | data source |
| [aws_kms_alias.secrets_manager_secret_encryption_key](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/kms_alias) | data source |
| [aws_region.current](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/region) | data source |
| [aws_secretsmanager_secret.gov_uk_notify_api_key](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/secretsmanager_secret) | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_allowed_api_arns"></a> [allowed\_api\_arns](#input\_allowed\_api\_arns) | n/a | `list(string)` | n/a | yes |
| <a name="input_app_public_url"></a> [app\_public\_url](#input\_app\_public\_url) | n/a | `string` | n/a | yes |
| <a name="input_event_bus_name"></a> [event\_bus\_name](#input\_event\_bus\_name) | n/a | `string` | n/a | yes |
| <a name="input_lambda_function_image_ecr_url"></a> [lambda\_function\_image\_ecr\_url](#input\_lambda\_function\_image\_ecr\_url) | n/a | `string` | n/a | yes |
| <a name="input_lambda_function_image_tag"></a> [lambda\_function\_image\_tag](#input\_lambda\_function\_image\_tag) | n/a | `string` | n/a | yes |
| <a name="input_lpas_table"></a> [lpas\_table](#input\_lpas\_table) | n/a | <pre>object({<br>    arn  = string<br>    name = string<br>  })</pre> | n/a | yes |
| <a name="input_uid_base_url"></a> [uid\_base\_url](#input\_uid\_base\_url) | n/a | `string` | n/a | yes |
| <a name="input_uploads_bucket"></a> [uploads\_bucket](#input\_uploads\_bucket) | n/a | `any` | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_lambda_function"></a> [lambda\_function](#output\_lambda\_function) | n/a |
<!-- END_TF_DOCS -->
