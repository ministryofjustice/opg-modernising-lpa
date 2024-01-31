# Global Resource Module

This module creates the global resources for an environment.

## Requirements

| Name                                                                      | Version   |
|---------------------------------------------------------------------------|-----------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.5.2  |
| <a name="requirement_aws"></a> [aws](#requirement\_aws)                   | ~> 5.34.0 |
| <a name="requirement_pagerduty"></a> [pagerduty](#requirement\_pagerduty) | 3.5.2     |

## Providers

| Name                                                                   | Version   |
|------------------------------------------------------------------------|-----------|
| <a name="provider_aws.global"></a> [aws.global](#provider\_aws.global) | ~> 5.34.0 |

## Modules

No modules.

## Resources

| Name                                                                                                                                                                  | Type        |
|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------|
| [aws_applicationinsights_application.environment_global](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/applicationinsights_application) | resource    |
| [aws_iam_role.app_task_role](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role)                                                    | resource    |
| [aws_iam_role.cross_account_put](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role)                                                | resource    |
| [aws_iam_role.execution_role](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role)                                                   | resource    |
| [aws_iam_role.s3_antivirus](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role)                                                     | resource    |
| [aws_iam_role_policy.execution_role](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role_policy)                                     | resource    |
| [aws_iam_role_policy_attachment.s3_antivirus_execution_role](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role_policy_attachment)  | resource    |
| [aws_resourcegroups_group.environment_global](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/resourcegroups_group)                       | resource    |
| [aws_default_tags.current](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/default_tags)                                               | data source |
| [aws_iam_policy_document.cross_account_put_assume_role](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document)           | data source |
| [aws_iam_policy_document.execution_role](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document)                          | data source |
| [aws_iam_policy_document.execution_role_assume_policy](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document)            | data source |
| [aws_iam_policy_document.lambda_assume](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document)                           | data source |
| [aws_iam_policy_document.task_role_assume_policy](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document)                 | data source |

## Inputs

| Name                                                                                                                                                          | Description                            | Type   | Default | Required |
|---------------------------------------------------------------------------------------------------------------------------------------------------------------|----------------------------------------|--------|---------|:--------:|
| <a name="input_cloudwatch_application_insights_enabled"></a> [cloudwatch\_application\_insights\_enabled](#input\_cloudwatch\_application\_insights\_enabled) | Enable CloudWatch Application Insights | `bool` | n/a     |   yes    |

## Outputs

| Name                                                                                           | Description |
|------------------------------------------------------------------------------------------------|-------------|
| <a name="output_iam_roles"></a> [iam\_roles](#output\_iam\_roles)                              | n/a         |
| <a name="output_resource_group_arn"></a> [resource\_group\_arn](#output\_resource\_group\_arn) | n/a         |

<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.5.2 |
| <a name="requirement_aws"></a> [aws](#requirement\_aws) | ~> 5.34.0 |
| <a name="requirement_pagerduty"></a> [pagerduty](#requirement\_pagerduty) | 3.5.2 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_aws.global"></a> [aws.global](#provider\_aws.global) | ~> 5.34.0 |

## Modules

No modules.

## Resources

| Name | Type |
|------|------|
| [aws_applicationinsights_application.environment_global](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/applicationinsights_application) | resource |
| [aws_iam_role.app_task_role](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role) | resource |
| [aws_iam_role.cross_account_put](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role) | resource |
| [aws_iam_role.execution_role](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role) | resource |
| [aws_iam_role.s3_antivirus](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role) | resource |
| [aws_iam_role_policy.execution_role](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role_policy) | resource |
| [aws_iam_role_policy_attachment.s3_antivirus_execution_role](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role_policy_attachment) | resource |
| [aws_resourcegroups_group.environment_global](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/resourcegroups_group) | resource |
| [aws_default_tags.current](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/default_tags) | data source |
| [aws_iam_policy_document.cross_account_put_assume_role](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_iam_policy_document.execution_role](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_iam_policy_document.execution_role_assume_policy](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_iam_policy_document.lambda_assume](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_iam_policy_document.task_role_assume_policy](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_cloudwatch_application_insights_enabled"></a> [cloudwatch\_application\_insights\_enabled](#input\_cloudwatch\_application\_insights\_enabled) | Enable CloudWatch Application Insights | `bool` | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_iam_roles"></a> [iam\_roles](#output\_iam\_roles) | n/a |
| <a name="output_resource_group_arn"></a> [resource\_group\_arn](#output\_resource\_group\_arn) | n/a |
<!-- END_TF_DOCS -->
