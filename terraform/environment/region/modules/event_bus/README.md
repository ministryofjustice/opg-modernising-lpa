# Reduced Fees events

This module creates event driven architecture to send and receive reduced fees events between MLPAB and Sirius.

Outbound events are sent to an Event Bridge event bus from DynamoDB Streams using EventBridge Pipes.

From there the event is sent to the Sirius event bus for processing using a rule with cross account putEvent IAM permissions.

You can create an incoming reduced fee approved event using the following aws cli put-events command (requires updating the `EventBusName` and the value of `uid` under `Detail`):

```shell
aws-vault exec mlpa-dev -- aws events put-events --entries file://reduced_fee_approved_event.json
```

<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.5.2 |
| <a name="requirement_aws"></a> [aws](#requirement\_aws) | ~> 5.42.0 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_aws"></a> [aws](#provider\_aws) | ~> 5.42.0 |
| <a name="provider_aws.region"></a> [aws.region](#provider\_aws.region) | ~> 5.42.0 |

## Modules

No modules.

## Resources

| Name | Type |
|------|------|
| [aws_cloudwatch_event_archive.main](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/cloudwatch_event_archive) | resource |
| [aws_cloudwatch_event_bus.main](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/cloudwatch_event_bus) | resource |
| [aws_cloudwatch_event_bus_policy.cross_account_receive](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/cloudwatch_event_bus_policy) | resource |
| [aws_cloudwatch_event_rule.cross_account_put](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/cloudwatch_event_rule) | resource |
| [aws_cloudwatch_event_target.cross_account_put](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/cloudwatch_event_target) | resource |
| [aws_cloudwatch_metric_alarm.event_bus_dead_letter_queue](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/cloudwatch_metric_alarm) | resource |
| [aws_iam_role_policy.cross_account_put](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role_policy) | resource |
| [aws_sqs_queue.event_bus_dead_letter_queue](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/sqs_queue) | resource |
| [aws_default_tags.current](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/default_tags) | data source |
| [aws_iam_policy_document.cross_account_put_access](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_iam_policy_document.cross_account_receive](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_iam_policy_document.sqs](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_kms_alias.sqs](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/kms_alias) | data source |
| [aws_region.current](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/region) | data source |
| [aws_sns_topic.custom_cloudwatch_alarms](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/sns_topic) | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_iam_role"></a> [iam\_role](#input\_iam\_role) | IAM role to allow cross account put to event bus | `any` | n/a | yes |
| <a name="input_receive_account_ids"></a> [receive\_account\_ids](#input\_receive\_account\_ids) | IDs of accounts to receive messages from | `list(string)` | `[]` | no |
| <a name="input_target_event_bus_arn"></a> [target\_event\_bus\_arn](#input\_target\_event\_bus\_arn) | ARN of the event bus to forward events to | `string` | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_event_bus"></a> [event\_bus](#output\_event\_bus) | n/a |
<!-- END_TF_DOCS -->
