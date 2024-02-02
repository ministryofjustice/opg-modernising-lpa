# Region Resources Module

This module creates the regional resources for an environment.

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
| <a name="provider_aws.management"></a> [aws.management](#provider\_aws.management) | ~> 5.34.0 |
| <a name="provider_aws.management_global"></a> [aws.management\_global](#provider\_aws.management\_global) | ~> 5.34.0 |
| <a name="provider_aws.region"></a> [aws.region](#provider\_aws.region) | ~> 5.34.0 |
| <a name="provider_pagerduty"></a> [pagerduty](#provider\_pagerduty) | 3.5.2 |

## Modules

| Name | Source | Version |
|------|--------|---------|
| <a name="module_app"></a> [app](#module\_app) | ./modules/app | n/a |
| <a name="module_app_ecs_autoscaling"></a> [app\_ecs\_autoscaling](#module\_app\_ecs\_autoscaling) | ./modules/ecs_autoscaling | n/a |
| <a name="module_application_logs"></a> [application\_logs](#module\_application\_logs) | ./modules/application_logs | n/a |
| <a name="module_event_bus"></a> [event\_bus](#module\_event\_bus) | ./modules/event_bus | n/a |
| <a name="module_event_received"></a> [event\_received](#module\_event\_received) | ./modules/event_received | n/a |
| <a name="module_mock_onelogin"></a> [mock\_onelogin](#module\_mock\_onelogin) | ./modules/mock_onelogin | n/a |
| <a name="module_s3_antivirus"></a> [s3\_antivirus](#module\_s3\_antivirus) | ./modules/s3_antivirus | n/a |
| <a name="module_uploads_s3_bucket"></a> [uploads\_s3\_bucket](#module\_uploads\_s3\_bucket) | ./modules/uploads_s3_bucket | n/a |

## Resources

| Name | Type |
|------|------|
| [aws_applicationinsights_application.environment](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/applicationinsights_application) | resource |
| [aws_cloudwatch_dashboard.health_checks](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/cloudwatch_dashboard) | resource |
| [aws_cloudwatch_metric_alarm.dependency_health_check](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/cloudwatch_metric_alarm) | resource |
| [aws_cloudwatch_metric_alarm.service_health_check](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/cloudwatch_metric_alarm) | resource |
| [aws_ecs_cluster.main](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/ecs_cluster) | resource |
| [aws_iam_role_policy.execution_role_region](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role_policy) | resource |
| [aws_iam_role_policy.rum_monitor_unauthenticated](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role_policy) | resource |
| [aws_resourcegroups_group.environment](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/resourcegroups_group) | resource |
| [aws_route53_health_check.dependency_health_check](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/route53_health_check) | resource |
| [aws_route53_health_check.service_health_check](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/route53_health_check) | resource |
| [aws_route53_record.app](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/route53_record) | resource |
| [aws_route53_record.mock_onelogin](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/route53_record) | resource |
| [aws_rum_app_monitor.main](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/rum_app_monitor) | resource |
| [aws_secretsmanager_secret.rum_monitor_application_id](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/secretsmanager_secret) | resource |
| [aws_secretsmanager_secret_version.rum_monitor_application_id](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/secretsmanager_secret_version) | resource |
| [aws_service_discovery_private_dns_namespace.mock_one_login](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/service_discovery_private_dns_namespace) | resource |
| [aws_sns_topic.dependency_health_checks_global](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/sns_topic) | resource |
| [aws_sns_topic.service_health_checks_global](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/sns_topic) | resource |
| [aws_sns_topic_subscription.cloudwatch_application_insights](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/sns_topic_subscription) | resource |
| [aws_sns_topic_subscription.dependency_health_check](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/sns_topic_subscription) | resource |
| [aws_sns_topic_subscription.ecs_autoscaling_alarms](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/sns_topic_subscription) | resource |
| [aws_sns_topic_subscription.service_health_check](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/sns_topic_subscription) | resource |
| [pagerduty_service_integration.cloudwatch_application_insights](https://registry.terraform.io/providers/PagerDuty/pagerduty/3.5.2/docs/resources/service_integration) | resource |
| [pagerduty_service_integration.dependency_health_check](https://registry.terraform.io/providers/PagerDuty/pagerduty/3.5.2/docs/resources/service_integration) | resource |
| [pagerduty_service_integration.ecs_autoscaling_alarms](https://registry.terraform.io/providers/PagerDuty/pagerduty/3.5.2/docs/resources/service_integration) | resource |
| [pagerduty_service_integration.service_health_check](https://registry.terraform.io/providers/PagerDuty/pagerduty/3.5.2/docs/resources/service_integration) | resource |
| [aws_availability_zones.available](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/availability_zones) | data source |
| [aws_caller_identity.current](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/caller_identity) | data source |
| [aws_default_tags.current](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/default_tags) | data source |
| [aws_ecr_image.s3_antivirus](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/ecr_image) | data source |
| [aws_ecr_repository.event_received](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/ecr_repository) | data source |
| [aws_ecr_repository.s3_antivirus](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/ecr_repository) | data source |
| [aws_ecr_repository.s3_create_batch_replication_jobs](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/ecr_repository) | data source |
| [aws_iam_policy_document.execution_role_region](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_iam_policy_document.rum_monitor_unauthenticated](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_iam_role.ecs_autoscaling_service_role](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_role) | data source |
| [aws_iam_role.rum_monitor_unauthenticated](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_role) | data source |
| [aws_iam_role.sns_failure_feedback](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_role) | data source |
| [aws_iam_role.sns_success_feedback](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_role) | data source |
| [aws_kms_alias.secrets_manager_secret_encryption_key](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/kms_alias) | data source |
| [aws_kms_alias.sns_kms_key_alias_global](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/kms_alias) | data source |
| [aws_region.current](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/region) | data source |
| [aws_route53_zone.modernising_lpa](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/route53_zone) | data source |
| [aws_s3_bucket.antivirus_definitions](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/s3_bucket) | data source |
| [aws_secretsmanager_secret_version.rum_monitor_identity_pool_id](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/secretsmanager_secret_version) | data source |
| [aws_sns_topic.cloudwatch_application_insights](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/sns_topic) | data source |
| [aws_sns_topic.custom_cloudwatch_alarms](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/sns_topic) | data source |
| [aws_sns_topic.ecs_autoscaling_alarms](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/sns_topic) | data source |
| [aws_ssm_parameter.additional_allowed_ingress_cidrs](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/ssm_parameter) | data source |
| [aws_ssm_parameter.replication_bucket_arn](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/ssm_parameter) | data source |
| [aws_ssm_parameter.replication_encryption_key](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/ssm_parameter) | data source |
| [aws_subnet.application](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/subnet) | data source |
| [aws_subnet.public](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/subnet) | data source |
| [aws_vpc.main](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/vpc) | data source |
| [pagerduty_service.main](https://registry.terraform.io/providers/PagerDuty/pagerduty/3.5.2/docs/data-sources/service) | data source |
| [pagerduty_vendor.cloudwatch](https://registry.terraform.io/providers/PagerDuty/pagerduty/3.5.2/docs/data-sources/vendor) | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_alb_deletion_protection_enabled"></a> [alb\_deletion\_protection\_enabled](#input\_alb\_deletion\_protection\_enabled) | If true, deletion of the load balancer will be disabled via the AWS API. This will prevent Terraform from deleting the load balancer. Defaults to false. | `bool` | n/a | yes |
| <a name="input_app_env_vars"></a> [app\_env\_vars](#input\_app\_env\_vars) | Environment variable values for app | `any` | n/a | yes |
| <a name="input_app_service_container_version"></a> [app\_service\_container\_version](#input\_app\_service\_container\_version) | Container version the app service | `string` | n/a | yes |
| <a name="input_app_service_repository_url"></a> [app\_service\_repository\_url](#input\_app\_service\_repository\_url) | Repository URL for the app service | `string` | n/a | yes |
| <a name="input_application_log_retention_days"></a> [application\_log\_retention\_days](#input\_application\_log\_retention\_days) | Specifies the number of days you want to retain log events in the specified log group. Possible values are: 1, 3, 5, 7, 14, 30, 60, 90, 120, 150, 180, 365, 400, 545, 731, 1827, 3653, and 0. If you select 0, the events in the log group are always retained and never expire. | `number` | n/a | yes |
| <a name="input_cloudwatch_application_insights_enabled"></a> [cloudwatch\_application\_insights\_enabled](#input\_cloudwatch\_application\_insights\_enabled) | Enable CloudWatch Application Insights | `bool` | n/a | yes |
| <a name="input_dependency_health_check_alarm_enabled"></a> [dependency\_health\_check\_alarm\_enabled](#input\_dependency\_health\_check\_alarm\_enabled) | Enable the dependency health check alert actions | `bool` | `false` | no |
| <a name="input_dns_weighting"></a> [dns\_weighting](#input\_dns\_weighting) | Weighting for DNS records | `number` | n/a | yes |
| <a name="input_ecs_capacity_provider"></a> [ecs\_capacity\_provider](#input\_ecs\_capacity\_provider) | Name of the capacity provider to use. Valid values are FARGATE\_SPOT and FARGATE | `string` | n/a | yes |
| <a name="input_ecs_task_autoscaling"></a> [ecs\_task\_autoscaling](#input\_ecs\_task\_autoscaling) | task minimum and maximum values for autoscaling | `any` | n/a | yes |
| <a name="input_iam_roles"></a> [iam\_roles](#input\_iam\_roles) | ARN of IAM role that allows your Amazon ECS container task to make calls to other AWS services. | <pre>object({<br>    ecs_execution_role        = any<br>    app_ecs_task_role         = any<br>    s3_antivirus              = any<br>    cross_account_put         = any<br>    fault_injection_simulator = any<br>  })</pre> | n/a | yes |
| <a name="input_ingress_allow_list_cidr"></a> [ingress\_allow\_list\_cidr](#input\_ingress\_allow\_list\_cidr) | List of CIDR ranges permitted to access the service | `list(string)` | n/a | yes |
| <a name="input_lpa_store_service"></a> [lpa\_store\_service](#input\_lpa\_store\_service) | n/a | <pre>object({<br>    base_url = string<br>    api_arns = list(string)<br>  })</pre> | n/a | yes |
| <a name="input_lpas_table"></a> [lpas\_table](#input\_lpas\_table) | DynamoDB table for storing LPAs | `any` | n/a | yes |
| <a name="input_mock_onelogin_enabled"></a> [mock\_onelogin\_enabled](#input\_mock\_onelogin\_enabled) | n/a | `bool` | n/a | yes |
| <a name="input_mock_onelogin_service_container_version"></a> [mock\_onelogin\_service\_container\_version](#input\_mock\_onelogin\_service\_container\_version) | Container version for the mock-onelogin service | `string` | n/a | yes |
| <a name="input_mock_onelogin_service_repository_url"></a> [mock\_onelogin\_service\_repository\_url](#input\_mock\_onelogin\_service\_repository\_url) | Repository URL for the mock-onelogin service | `string` | n/a | yes |
| <a name="input_pagerduty_service_name"></a> [pagerduty\_service\_name](#input\_pagerduty\_service\_name) | Name of the PagerDuty service to use for alerts | `string` | n/a | yes |
| <a name="input_public_access_enabled"></a> [public\_access\_enabled](#input\_public\_access\_enabled) | Enable access to the Modernising LPA service from the public internet | `bool` | n/a | yes |
| <a name="input_receive_account_ids"></a> [receive\_account\_ids](#input\_receive\_account\_ids) | IDs of accounts to receive messages from | `list(string)` | `[]` | no |
| <a name="input_reduced_fees"></a> [reduced\_fees](#input\_reduced\_fees) | n/a | <pre>object({<br>    s3_object_replication_enabled             = bool<br>    target_environment                        = string<br>    destination_account_id                    = string<br>    enable_s3_batch_job_replication_scheduler = bool<br>  })</pre> | n/a | yes |
| <a name="input_s3_antivirus_provisioned_concurrency"></a> [s3\_antivirus\_provisioned\_concurrency](#input\_s3\_antivirus\_provisioned\_concurrency) | Number of concurrent executions to provision for Lambda | `number` | `0` | no |
| <a name="input_service_health_check_alarm_enabled"></a> [service\_health\_check\_alarm\_enabled](#input\_service\_health\_check\_alarm\_enabled) | Enable the service health check alert actions | `bool` | `false` | no |
| <a name="input_target_event_bus_arn"></a> [target\_event\_bus\_arn](#input\_target\_event\_bus\_arn) | ARN of the event bus to forward events to | `string` | n/a | yes |
| <a name="input_uid_service"></a> [uid\_service](#input\_uid\_service) | n/a | <pre>object({<br>    base_url = string<br>    api_arns = list(string)<br>  })</pre> | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_app_fqdn"></a> [app\_fqdn](#output\_app\_fqdn) | n/a |
| <a name="output_app_load_balancer"></a> [app\_load\_balancer](#output\_app\_load\_balancer) | n/a |
| <a name="output_app_load_balancer_security_group"></a> [app\_load\_balancer\_security\_group](#output\_app\_load\_balancer\_security\_group) | n/a |
| <a name="output_resource_group_arn"></a> [resource\_group\_arn](#output\_resource\_group\_arn) | n/a |
<!-- END_TF_DOCS -->
