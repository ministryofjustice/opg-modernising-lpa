# App Module

The module creates an ECS service for the Modernising LPA application, and associated resources including a load balancer, security groups, and a WAFv2 web ACL association.

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
| [aws_ecs_service.app](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/ecs_service) | resource |
| [aws_ecs_task_definition.app](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/ecs_task_definition) | resource |
| [aws_iam_role_policy.app_task_role](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role_policy) | resource |
| [aws_lb.app](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/lb) | resource |
| [aws_lb_listener.app_loadbalancer](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/lb_listener) | resource |
| [aws_lb_listener.app_loadbalancer_http_redirect](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/lb_listener) | resource |
| [aws_lb_listener_certificate.app_loadbalancer_live_service_certificate](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/lb_listener_certificate) | resource |
| [aws_lb_listener_rule.app_maintenance](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/lb_listener_rule) | resource |
| [aws_lb_listener_rule.app_maintenance_welsh](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/lb_listener_rule) | resource |
| [aws_lb_target_group.app](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/lb_target_group) | resource |
| [aws_security_group.app_ecs_service](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/security_group) | resource |
| [aws_security_group.app_loadbalancer](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/security_group) | resource |
| [aws_security_group_rule.app_ecs_service_egress](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/security_group_rule) | resource |
| [aws_security_group_rule.app_ecs_service_ingress](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/security_group_rule) | resource |
| [aws_security_group_rule.app_loadbalancer_egress](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/security_group_rule) | resource |
| [aws_security_group_rule.app_loadbalancer_ingress](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/security_group_rule) | resource |
| [aws_security_group_rule.app_loadbalancer_port_80_redirect_ingress](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/security_group_rule) | resource |
| [aws_security_group_rule.app_loadbalancer_public_access_ingress](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/security_group_rule) | resource |
| [aws_security_group_rule.app_loadbalancer_public_access_ingress_port_80](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/security_group_rule) | resource |
| [aws_security_group_rule.loadbalancer_ingress_route53_healthchecks](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/security_group_rule) | resource |
| [aws_wafv2_web_acl_association.app](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/wafv2_web_acl_association) | resource |
| [aws_acm_certificate.certificate_app](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/acm_certificate) | data source |
| [aws_default_tags.current](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/default_tags) | data source |
| [aws_iam_policy_document.combined](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_iam_policy_document.ecs_task_role_fis_related_task_permissions](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_iam_policy_document.task_role_access_policy](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_ip_ranges.route53_healthchecks](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/ip_ranges) | data source |
| [aws_kms_alias.dynamodb_encryption_key](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/kms_alias) | data source |
| [aws_kms_alias.opensearch_encryption_key](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/kms_alias) | data source |
| [aws_kms_alias.reduced_fees_uploads_s3_encryption](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/kms_alias) | data source |
| [aws_kms_alias.secrets_manager_secret_encryption_key](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/kms_alias) | data source |
| [aws_region.current](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/region) | data source |
| [aws_s3_bucket.access_log](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/s3_bucket) | data source |
| [aws_secretsmanager_secret.cookie_session_keys](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/secretsmanager_secret) | data source |
| [aws_secretsmanager_secret.gov_uk_notify_api_key](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/secretsmanager_secret) | data source |
| [aws_secretsmanager_secret.gov_uk_onelogin_identity_public_key](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/secretsmanager_secret) | data source |
| [aws_secretsmanager_secret.gov_uk_pay_api_key](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/secretsmanager_secret) | data source |
| [aws_secretsmanager_secret.lpa_store_jwt_secret_key](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/secretsmanager_secret) | data source |
| [aws_secretsmanager_secret.os_postcode_lookup_api_key](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/secretsmanager_secret) | data source |
| [aws_secretsmanager_secret.private_jwt_key](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/secretsmanager_secret) | data source |
| [aws_secretsmanager_secret.rum_monitor_identity_pool_id](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/secretsmanager_secret) | data source |
| [aws_wafv2_web_acl.main](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/wafv2_web_acl) | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_alb_deletion_protection_enabled"></a> [alb\_deletion\_protection\_enabled](#input\_alb\_deletion\_protection\_enabled) | If true, deletion of the load balancer will be disabled via the AWS API. This will prevent Terraform from deleting the load balancer. Defaults to false. | `bool` | n/a | yes |
| <a name="input_app_allowed_api_arns"></a> [app\_allowed\_api\_arns](#input\_app\_allowed\_api\_arns) | n/a | `list(string)` | n/a | yes |
| <a name="input_app_env_vars"></a> [app\_env\_vars](#input\_app\_env\_vars) | Environment variable values for app | `any` | n/a | yes |
| <a name="input_app_service_container_version"></a> [app\_service\_container\_version](#input\_app\_service\_container\_version) | (optional) describe your variable | `string` | n/a | yes |
| <a name="input_app_service_repository_url"></a> [app\_service\_repository\_url](#input\_app\_service\_repository\_url) | (optional) describe your variable | `string` | n/a | yes |
| <a name="input_aws_rum_guest_role_arn"></a> [aws\_rum\_guest\_role\_arn](#input\_aws\_rum\_guest\_role\_arn) | ARN of the AWS RUM guest role | `string` | n/a | yes |
| <a name="input_container_port"></a> [container\_port](#input\_container\_port) | Port on the container to associate with. | `number` | n/a | yes |
| <a name="input_ecs_application_log_group_name"></a> [ecs\_application\_log\_group\_name](#input\_ecs\_application\_log\_group\_name) | The AWS Cloudwatch Log Group resource for application logging | `string` | n/a | yes |
| <a name="input_ecs_capacity_provider"></a> [ecs\_capacity\_provider](#input\_ecs\_capacity\_provider) | Name of the capacity provider to use. Valid values are FARGATE\_SPOT and FARGATE | `string` | n/a | yes |
| <a name="input_ecs_cluster"></a> [ecs\_cluster](#input\_ecs\_cluster) | ARN of an ECS cluster. | `string` | n/a | yes |
| <a name="input_ecs_execution_role"></a> [ecs\_execution\_role](#input\_ecs\_execution\_role) | ID and ARN of the task execution role that the Amazon ECS container agent and the Docker daemon can assume. | <pre>object({<br>    id  = string<br>    arn = string<br>  })</pre> | n/a | yes |
| <a name="input_ecs_service_desired_count"></a> [ecs\_service\_desired\_count](#input\_ecs\_service\_desired\_count) | Number of instances of the task definition to place and keep running. Defaults to 0. Do not specify if using the DAEMON scheduling strategy. | `number` | `0` | no |
| <a name="input_ecs_task_role"></a> [ecs\_task\_role](#input\_ecs\_task\_role) | ARN of IAM role that allows your Amazon ECS container task to make calls to other AWS services. | `any` | n/a | yes |
| <a name="input_event_bus"></a> [event\_bus](#input\_event\_bus) | Name and ARN of the event bus to send events to | <pre>object({<br>    name = string<br>    arn  = string<br>  })</pre> | n/a | yes |
| <a name="input_fault_injection_experiments_enabled"></a> [fault\_injection\_experiments\_enabled](#input\_fault\_injection\_experiments\_enabled) | Enable fault injection | `bool` | n/a | yes |
| <a name="input_ingress_allow_list_cidr"></a> [ingress\_allow\_list\_cidr](#input\_ingress\_allow\_list\_cidr) | List of CIDR ranges permitted to access the service | `list(string)` | n/a | yes |
| <a name="input_lpa_store_base_url"></a> [lpa\_store\_base\_url](#input\_lpa\_store\_base\_url) | n/a | `string` | n/a | yes |
| <a name="input_lpas_table"></a> [lpas\_table](#input\_lpas\_table) | DynamoDB table for storing LPAs | `any` | n/a | yes |
| <a name="input_mock_onelogin_enabled"></a> [mock\_onelogin\_enabled](#input\_mock\_onelogin\_enabled) | n/a | `bool` | n/a | yes |
| <a name="input_network"></a> [network](#input\_network) | VPC ID, a list of application subnets, and a list of private subnets required to provision the ECS service | <pre>object({<br>    vpc_id              = string<br>    application_subnets = list(string)<br>    public_subnets      = list(string)<br>  })</pre> | n/a | yes |
| <a name="input_public_access_enabled"></a> [public\_access\_enabled](#input\_public\_access\_enabled) | Enable access to the Modernising LPA service from the public internet | `bool` | n/a | yes |
| <a name="input_rum_monitor_application_id_secretsmanager_secret_arn"></a> [rum\_monitor\_application\_id\_secretsmanager\_secret\_arn](#input\_rum\_monitor\_application\_id\_secretsmanager\_secret\_arn) | ARN of the AWS Secrets Manager secret containing the RUM monitor application ID | `string` | n/a | yes |
| <a name="input_search_collection_arn"></a> [search\_collection\_arn](#input\_search\_collection\_arn) | ARN of the OpenSearch collection to use | `string` | n/a | yes |
| <a name="input_search_endpoint"></a> [search\_endpoint](#input\_search\_endpoint) | URL of the OpenSearch Service endpoint to use | `string` | n/a | yes |
| <a name="input_uid_base_url"></a> [uid\_base\_url](#input\_uid\_base\_url) | n/a | `string` | n/a | yes |
| <a name="input_uploads_s3_bucket"></a> [uploads\_s3\_bucket](#input\_uploads\_s3\_bucket) | Name and ARN of the S3 bucket for uploads | <pre>object({<br>    bucket_name = string<br>    bucket_arn  = string<br>  })</pre> | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_ecs_service"></a> [ecs\_service](#output\_ecs\_service) | n/a |
| <a name="output_ecs_service_security_group"></a> [ecs\_service\_security\_group](#output\_ecs\_service\_security\_group) | n/a |
| <a name="output_load_balancer"></a> [load\_balancer](#output\_load\_balancer) | n/a |
| <a name="output_load_balancer_security_group"></a> [load\_balancer\_security\_group](#output\_load\_balancer\_security\_group) | n/a |
<!-- END_TF_DOCS -->
