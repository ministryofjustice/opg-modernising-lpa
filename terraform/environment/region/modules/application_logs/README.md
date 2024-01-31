# Application Logs

The module creates a cloudwatch log group and useful log queries for application logs.

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

| Name                                                                                                                                                              | Type        |
|-------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------|
| [aws_cloudwatch_log_group.application_logs](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/cloudwatch_log_group)                     | resource    |
| [aws_cloudwatch_query_definition.app_container_messages](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/cloudwatch_query_definition) | resource    |
| [aws_default_tags.current](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/default_tags)                                           | data source |
| [aws_kms_alias.cloudwatch_application_logs_encryption](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/kms_alias)                  | data source |

## Inputs

| Name                                                                                                                               | Description                                                                                                                                                                                                                                                                      | Type     | Default | Required |
|------------------------------------------------------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|----------|---------|:--------:|
| <a name="input_application_log_retention_days"></a> [application\_log\_retention\_days](#input\_application\_log\_retention\_days) | Specifies the number of days you want to retain log events in the specified log group. Possible values are: 1, 3, 5, 7, 14, 30, 60, 90, 120, 150, 180, 365, 400, 545, 731, 1827, 3653, and 0. If you select 0, the events in the log group are always retained and never expire. | `number` | n/a     |   yes    |

## Outputs

| Name                                                                                                 | Description |
|------------------------------------------------------------------------------------------------------|-------------|
| <a name="output_cloudwatch_log_group"></a> [cloudwatch\_log\_group](#output\_cloudwatch\_log\_group) | n/a         |
