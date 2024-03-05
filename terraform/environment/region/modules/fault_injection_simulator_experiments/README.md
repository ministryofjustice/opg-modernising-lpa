<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.5.2 |
| <a name="requirement_aws"></a> [aws](#requirement\_aws) | ~> 5.39.0 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_aws.region"></a> [aws.region](#provider\_aws.region) | ~> 5.39.0 |

## Modules

No modules.

## Resources

| Name | Type |
|------|------|
| [aws_cloudwatch_log_group.fis_app_ecs_tasks](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/cloudwatch_log_group) | resource |
| [aws_cloudwatch_log_resource_policy.fis_app_ecs_tasks](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/cloudwatch_log_resource_policy) | resource |
| [aws_fis_experiment_template.ecs_app_cpu_stress](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/fis_experiment_template) | resource |
| [aws_fis_experiment_template.ecs_app_io_stress](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/fis_experiment_template) | resource |
| [aws_fis_experiment_template.ecs_app_stop_tasks](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/fis_experiment_template) | resource |
| [aws_iam_role_policy.fis_role_log_encryption](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role_policy) | resource |
| [aws_caller_identity.current](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/caller_identity) | data source |
| [aws_default_tags.current](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/default_tags) | data source |
| [aws_iam_policy_document.cloudwatch_log_group_policy_fis_app_ecs_tasks](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_iam_policy_document.fis_role_log_encryption](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_kms_alias.cloudwatch_application_logs_encryption](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/kms_alias) | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_ecs_cluster"></a> [ecs\_cluster](#input\_ecs\_cluster) | Name of the ECS cluster to run the experiments on. | `string` | n/a | yes |
| <a name="input_fault_injection_simulator_role"></a> [fault\_injection\_simulator\_role](#input\_fault\_injection\_simulator\_role) | ARN of IAM role that allows AWS FIS to make calls to other AWS services. | `any` | n/a | yes |

## Outputs

No outputs.
<!-- END_TF_DOCS -->
