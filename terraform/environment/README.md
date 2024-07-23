# Terraform Shared

This terraform configuration manages per-environment resources.

just a random change

Per-account or otherwise shared resources are managed in `../account`

## Namespace resources

It is important to namespace resources to avoid getting errors for creating resources that already exist.

There are two namespace variables available.

```hcl
"${local.default_tags.environment-name}"
```

is available in the root module. Within modules, we use the default tags data source

```hcl
data "aws_default_tags" "current" {
  provider = aws.region
}

"${data.aws_default_tags.current.tags["environment-name"]}"
```

They will both return values like `1015mlpab17` or `production`

## Regional Design Pattern

The design intent for this project is to prepare infrastructure that can be replicated across regions, sharing global resources between them.

```shell
.
├── region
│   ├── modules
│   │   └── app
│   │       ├── ecs.tf
│   │       ├── alb.tf
│   │       └── terraform.tf
│   ├── app.tf
│   ├── network.tf
│   ├── terraform.tf
│   └── variables.tf
├── README.md
├── regions.tf
├── terraform.tf
```

Regions.tf will instantiate the /region module for each AWS region required.

Resources inside /region will be grouped as modules also, allowing for parts of a region to be replicated as and when needed.

This will allow us to deploy the service in a way that is globally resiliant, and highly available.

## Running Terraform Locally

This repository comes with an `.envrc` file containing useful environment variables for working with this repository.

`.envrc` can be sourced automatically using either [direnv](https://direnv.net) or manually with bash.

```shell
source .envrc
```

```shell
direnv allow
```

This sets environment variables that allow the following commands with no further setup

```shell
aws-vault exec identity -- terraform init
aws-vault exec identity -- terraform plan
aws-vault exec identity -- terraform force-unlock 49b3784c-51eb-668d-ac4b-3bd5b8701925
```

## Fixing state lock issue

A Terraform state lock error can happen if a terraform job is forcefully terminated (normal ctrl+c gracefully releases state lock).

CircleCI terminates a process if you cancel a job, so state lock doesn't get released.

Here's how to fix it if it happens.
Error:

```shell
rror locking state: Error acquiring the state lock: ConditionalCheckFailedException: The conditional request failed
    status code: 400, request id: 60Q304F4TMIRB13AMS36M49ND7VV4KQNSO5AEMVJF66Q9ASUAAJG
Lock Info:
  ID:        69592de7-6132-c863-ae53-976776ffe6cf
  Path:      opg.terraform.state/env:/development/opg-modernising-lpa/terraform.tfstate
  Operation: OperationTypeApply
  Who:       @d701fcddc381
  Version:   0.11.13
  Created:   2019-05-09 16:01:50.027392879 +0000 UTC
  Info:
```

Fix:

```shell
aws-vault exec identity -- terraform init
aws-vault exec identity -- terraform workspace select development
aws-vault exec identity -- terraform force-unlock 69592de7-6132-c863-ae53-976776ffe6cf
```

It is important to select the correct workspace.
For terraform_environment, this will be based on your PR and can be found in the Github Actions pipeline job `PR Environment Deploy`

<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | 1.7.5 |
| <a name="requirement_aws"></a> [aws](#requirement\_aws) | ~> 5.42.0 |
| <a name="requirement_pagerduty"></a> [pagerduty](#requirement\_pagerduty) | 3.10.0 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_aws"></a> [aws](#provider\_aws) | 5.42.0 |
| <a name="provider_aws.eu_west_1"></a> [aws.eu\_west\_1](#provider\_aws.eu\_west\_1) | 5.42.0 |
| <a name="provider_aws.eu_west_2"></a> [aws.eu\_west\_2](#provider\_aws.eu\_west\_2) | 5.42.0 |
| <a name="provider_aws.global"></a> [aws.global](#provider\_aws.global) | 5.42.0 |
| <a name="provider_aws.management_eu_west_1"></a> [aws.management\_eu\_west\_1](#provider\_aws.management\_eu\_west\_1) | 5.42.0 |
| <a name="provider_aws.management_global"></a> [aws.management\_global](#provider\_aws.management\_global) | 5.42.0 |
| <a name="provider_pagerduty"></a> [pagerduty](#provider\_pagerduty) | 3.10.0 |

## Modules

| Name | Source | Version |
|------|--------|---------|
| <a name="module_allow_list"></a> [allow\_list](#module\_allow\_list) | git@github.com:ministryofjustice/opg-terraform-aws-moj-ip-allow-list.git | v3.0.1 |
| <a name="module_eu_west_1"></a> [eu\_west\_1](#module\_eu\_west\_1) | ./region | n/a |
| <a name="module_eu_west_2"></a> [eu\_west\_2](#module\_eu\_west\_2) | ./region | n/a |
| <a name="module_global"></a> [global](#module\_global) | ./global | n/a |

## Resources

| Name | Type |
|------|------|
| [aws_backup_plan.main](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/backup_plan) | resource |
| [aws_backup_selection.main](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/backup_selection) | resource |
| [aws_backup_vault_notifications.aws_backup_failure_events](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/backup_vault_notifications) | resource |
| [aws_cloudwatch_log_group.opensearch_pipeline](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/cloudwatch_log_group) | resource |
| [aws_cloudwatch_metric_alarm.opensearch_4xx_errors](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/cloudwatch_metric_alarm) | resource |
| [aws_cloudwatch_metric_alarm.opensearch_5xx_errors](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/cloudwatch_metric_alarm) | resource |
| [aws_cloudwatch_query_definition.opensearch_pipeline](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/cloudwatch_query_definition) | resource |
| [aws_dynamodb_table.lpas_table](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/dynamodb_table) | resource |
| [aws_dynamodb_table_replica.lpas_table](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/dynamodb_table_replica) | resource |
| [aws_iam_role_policy.opensearch_pipeline](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role_policy) | resource |
| [aws_opensearchserverless_access_policy.app](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/opensearchserverless_access_policy) | resource |
| [aws_opensearchserverless_access_policy.event_received](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/opensearchserverless_access_policy) | resource |
| [aws_opensearchserverless_access_policy.pipeline](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/opensearchserverless_access_policy) | resource |
| [aws_opensearchserverless_access_policy.team_breakglas_access](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/opensearchserverless_access_policy) | resource |
| [aws_opensearchserverless_access_policy.team_operator_access](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/opensearchserverless_access_policy) | resource |
| [aws_opensearchserverless_collection.lpas_collection](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/opensearchserverless_collection) | resource |
| [aws_opensearchserverless_security_policy.lpas_collection_encryption_policy](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/opensearchserverless_security_policy) | resource |
| [aws_opensearchserverless_security_policy.lpas_collection_network_policy](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/opensearchserverless_security_policy) | resource |
| [aws_osis_pipeline.lpas](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/osis_pipeline) | resource |
| [aws_s3_bucket.opensearch_pipeline](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/s3_bucket) | resource |
| [aws_security_group.opensearch_ingestion](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/security_group) | resource |
| [aws_sns_topic.aws_backup_failure_events](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/sns_topic) | resource |
| [aws_sns_topic.opensearch](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/sns_topic) | resource |
| [aws_sns_topic_policy.aws_backup_failure_events](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/sns_topic_policy) | resource |
| [aws_sns_topic_subscription.opensearch](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/sns_topic_subscription) | resource |
| [aws_ssm_parameter.container_version](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/ssm_parameter) | resource |
| [aws_ssm_parameter.dns_target_region](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/ssm_parameter) | resource |
| [pagerduty_service_integration.opensearch](https://registry.terraform.io/providers/PagerDuty/pagerduty/3.10.0/docs/resources/service_integration) | resource |
| [aws_availability_zones.available](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/availability_zones) | data source |
| [aws_backup_vault.eu_west_1](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/backup_vault) | data source |
| [aws_backup_vault.eu_west_2](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/backup_vault) | data source |
| [aws_caller_identity.eu_west_1](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/caller_identity) | data source |
| [aws_ecr_repository.app](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/ecr_repository) | data source |
| [aws_ecr_repository.mock_onelogin](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/ecr_repository) | data source |
| [aws_iam_policy_document.aws_backup_sns](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_iam_policy_document.opensearch_pipeline](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_iam_role.aws_backup_role](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_role) | data source |
| [aws_iam_role.sns_failure_feedback](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_role) | data source |
| [aws_iam_role.sns_success_feedback](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_role) | data source |
| [aws_kms_alias.dynamodb_encryption_key](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/kms_alias) | data source |
| [aws_kms_alias.dynamodb_encryption_key_eu_west_1](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/kms_alias) | data source |
| [aws_kms_alias.dynamodb_encryption_key_eu_west_2](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/kms_alias) | data source |
| [aws_kms_alias.opensearch](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/kms_alias) | data source |
| [aws_kms_alias.opensearch_encryption_key](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/kms_alias) | data source |
| [aws_kms_alias.sns_encryption_key_eu_west_1](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/kms_alias) | data source |
| [aws_kms_alias.sns_kms_key_alias](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/kms_alias) | data source |
| [aws_subnet.application](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/subnet) | data source |
| [aws_vpc.main](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/vpc) | data source |
| [aws_vpc_endpoint.opensearch](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/vpc_endpoint) | data source |
| [pagerduty_service.main](https://registry.terraform.io/providers/PagerDuty/pagerduty/3.10.0/docs/data-sources/service) | data source |
| [pagerduty_vendor.cloudwatch](https://registry.terraform.io/providers/PagerDuty/pagerduty/3.10.0/docs/data-sources/vendor) | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_container_version"></a> [container\_version](#input\_container\_version) | n/a | `string` | `"latest"` | no |
| <a name="input_default_role"></a> [default\_role](#input\_default\_role) | n/a | `string` | `"modernising-lpa-ci"` | no |
| <a name="input_environments"></a> [environments](#input\_environments) | n/a | <pre>map(<br>    object({<br>      account_id    = string<br>      account_name  = string<br>      is_production = bool<br>      regions       = list(string)<br>      app = object({<br>        env = object({<br>          app_public_url         = string<br>          auth_redirect_base_url = string<br>          notify_is_production   = string<br>          onelogin_url           = string<br>        })<br>        autoscaling = object({<br>          minimum = number<br>          maximum = number<br>        })<br>        dependency_health_check_alarm_enabled   = bool<br>        service_health_check_alarm_enabled      = bool<br>        cloudwatch_application_insights_enabled = bool<br>        fault_injection_experiments_enabled     = bool<br>        real_user_monitoring_cw_logs_enabled    = bool<br>      })<br>      mock_onelogin_enabled = bool<br>      uid_service = object({<br>        base_url = string<br>        api_arns = list(string)<br>      })<br>      lpa_store_service = object({<br>        base_url = string<br>        api_arns = list(string)<br>      })<br>      backups = object({<br>        backup_plan_enabled = bool<br>        copy_action_enabled = bool<br>      })<br>      dynamodb = object({<br>        region_replica_enabled = bool<br>        stream_enabled         = bool<br>      })<br>      ecs = object({<br>        fargate_spot_capacity_provider_enabled = bool<br><br>      })<br>      cloudwatch_log_groups = object({<br>        application_log_retention_days = number<br>      })<br>      application_load_balancer = object({<br>        deletion_protection_enabled = bool<br>      })<br>      cloudwatch_application_insights_enabled = bool<br>      pagerduty_service_name                  = string<br>      event_bus = object({<br>        target_event_bus_arn = string<br>        receive_account_ids  = list(string)<br>      })<br>      reduced_fees = object({<br>        enabled                                   = bool<br>        s3_object_replication_enabled             = bool<br>        target_environment                        = string<br>        destination_account_id                    = string<br>        enable_s3_batch_job_replication_scheduler = bool<br>      })<br>      s3_antivirus_provisioned_concurrency = number<br>    })<br>  )</pre> | n/a | yes |
| <a name="input_pagerduty_api_key"></a> [pagerduty\_api\_key](#input\_pagerduty\_api\_key) | n/a | `string` | n/a | yes |
| <a name="input_public_access_enabled"></a> [public\_access\_enabled](#input\_public\_access\_enabled) | n/a | `bool` | `false` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_app_fqdn"></a> [app\_fqdn](#output\_app\_fqdn) | n/a |
| <a name="output_container_version"></a> [container\_version](#output\_container\_version) | n/a |
| <a name="output_environment_config_json"></a> [environment\_config\_json](#output\_environment\_config\_json) | n/a |
| <a name="output_public_access_enabled"></a> [public\_access\_enabled](#output\_public\_access\_enabled) | n/a |
| <a name="output_resource_group_arns"></a> [resource\_group\_arns](#output\_resource\_group\_arns) | n/a |
| <a name="output_workspace_name"></a> [workspace\_name](#output\_workspace\_name) | n/a |
<!-- END_TF_DOCS -->
