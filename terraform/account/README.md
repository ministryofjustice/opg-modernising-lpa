# Terraform Shared

This terraform configuration manages per-account or otherwise shared resources.

Per-environment resources are managed in ../environment

## Running Terraform Locally

This repository comes with an `.envrc` file containing useful environment variables for working with this repository.

`.envrc` can be sourced automatically using either [direnv](https://direnv.net) or manually with bash.

```bash
source .envrc
```

```bash
direnv allow
```

This sets environment variables that allow the following commands with no further setup needed.

```bash
aws-vault exec identity -- terraform init
aws-vault exec identity -- terraform plan
aws-vault exec identity -- terraform force-unlock 49b3784c-51eb-668d-ac4b-3bd5b8701925
```

## Regional Design Pattern

The design intent for this project is to prepare infrastructure that can be replicated across regions, sharing global resources between them.

```shell
.
├── region
│   ├── modules
│   │   └── certificates
│   │       ├── main.tf
│   │       └── terraform.tf
│   ├── certificates.tf
│   ├── network.tf
│   ├── terraform.tf
│   └── variables.tf
├── README.md
├── regions.tf
├── terraform.tf
```

Regions.tf will instatiate the /region module for each AWS region required.

Rsources inside /region will be grouped as modules also, allowing for parts of a region to be replicated as and when needed.

This will allow us to deploy the service in a way that is globally resiliant, and highly available.

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
For terraform_environment, this will be based on your PR and can be found in the Github Actions account level plan/apply pipeline job for example `TF Plan Dev Account`

<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | 1.7.2 |
| <a name="requirement_aws"></a> [aws](#requirement\_aws) | ~> 5.35.0 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_aws.eu_west_1"></a> [aws.eu\_west\_1](#provider\_aws.eu\_west\_1) | 5.35.0 |
| <a name="provider_aws.eu_west_2"></a> [aws.eu\_west\_2](#provider\_aws.eu\_west\_2) | 5.35.0 |
| <a name="provider_aws.global"></a> [aws.global](#provider\_aws.global) | 5.35.0 |
| <a name="provider_aws.management_global"></a> [aws.management\_global](#provider\_aws.management\_global) | 5.35.0 |

## Modules

| Name | Source | Version |
|------|--------|---------|
| <a name="module_eu_west_1"></a> [eu\_west\_1](#module\_eu\_west\_1) | ./region | n/a |
| <a name="module_eu_west_2"></a> [eu\_west\_2](#module\_eu\_west\_2) | ./region | n/a |

## Resources

| Name | Type |
|------|------|
| [aws_backup_vault.eu_west_1](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/backup_vault) | resource |
| [aws_backup_vault.eu_west_2](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/backup_vault) | resource |
| [aws_dynamodb_table.workspace_cleanup_table](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/dynamodb_table) | resource |
| [aws_iam_role.aws_backup_role](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role) | resource |
| [aws_iam_role.replication](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role) | resource |
| [aws_iam_role_policy_attachment.aws_backup_role](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role_policy_attachment) | resource |
| [aws_iam_service_linked_role.ecs_autoscaling_service_role](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_service_linked_role) | resource |
| [aws_kms_alias.cloudwatch_alias_eu_west_1](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kms_alias) | resource |
| [aws_kms_alias.cloudwatch_alias_eu_west_2](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kms_alias) | resource |
| [aws_kms_alias.dynamodb_alias_eu_west_1](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kms_alias) | resource |
| [aws_kms_alias.dynamodb_alias_eu_west_2](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kms_alias) | resource |
| [aws_kms_alias.reduced_fees_uploads_s3_alias_eu_west_1](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kms_alias) | resource |
| [aws_kms_alias.reduced_fees_uploads_s3_alias_eu_west_2](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kms_alias) | resource |
| [aws_kms_alias.secrets_manager_alias_eu_west_1](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kms_alias) | resource |
| [aws_kms_alias.secrets_manager_alias_eu_west_2](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kms_alias) | resource |
| [aws_kms_alias.sns_alias_eu_west_1](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kms_alias) | resource |
| [aws_kms_alias.sns_alias_eu_west_2](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kms_alias) | resource |
| [aws_kms_alias.sns_alias_global](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kms_alias) | resource |
| [aws_kms_alias.sqs_alias_eu_west_1](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kms_alias) | resource |
| [aws_kms_alias.sqs_alias_eu_west_2](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kms_alias) | resource |
| [aws_kms_key.cloudwatch](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kms_key) | resource |
| [aws_kms_key.dynamodb](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kms_key) | resource |
| [aws_kms_key.reduced_fees_uploads_s3](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kms_key) | resource |
| [aws_kms_key.secrets_manager](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kms_key) | resource |
| [aws_kms_key.sns](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kms_key) | resource |
| [aws_kms_key.sqs](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kms_key) | resource |
| [aws_kms_replica_key.cloudwatch_replica](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kms_replica_key) | resource |
| [aws_kms_replica_key.dynamodb_replica](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kms_replica_key) | resource |
| [aws_kms_replica_key.reduced_fees_uploads_s3_replica](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kms_replica_key) | resource |
| [aws_kms_replica_key.secrets_manager_replica](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kms_replica_key) | resource |
| [aws_kms_replica_key.sns_replica](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kms_replica_key) | resource |
| [aws_kms_replica_key.sns_replica_global](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kms_replica_key) | resource |
| [aws_kms_replica_key.sqs_replica](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kms_replica_key) | resource |
| [aws_resourcegroups_group.account_eu_west_1](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/resourcegroups_group) | resource |
| [aws_resourcegroups_group.account_eu_west_2](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/resourcegroups_group) | resource |
| [aws_resourcegroups_group.account_global](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/resourcegroups_group) | resource |
| [aws_secretsmanager_secret.cookie_session_keys](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/secretsmanager_secret) | resource |
| [aws_secretsmanager_secret.gov_uk_notify_api_key](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/secretsmanager_secret) | resource |
| [aws_secretsmanager_secret.gov_uk_onelogin_identity_public_key](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/secretsmanager_secret) | resource |
| [aws_secretsmanager_secret.gov_uk_pay_api_key](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/secretsmanager_secret) | resource |
| [aws_secretsmanager_secret.lpa_store_jwt_secret_key](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/secretsmanager_secret) | resource |
| [aws_secretsmanager_secret.os_postcode_lookup_api_key](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/secretsmanager_secret) | resource |
| [aws_secretsmanager_secret.private_jwt_key_base64](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/secretsmanager_secret) | resource |
| [aws_ssm_parameter.additional_allowed_ingress](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/ssm_parameter) | resource |
| [aws_caller_identity.global](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/caller_identity) | data source |
| [aws_default_tags.global](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/default_tags) | data source |
| [aws_iam_policy_document.assume_replication_role](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_iam_policy_document.aws_backup_assume_policy](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_iam_policy_document.cloudwatch_kms](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_iam_policy_document.cloudwatch_kms_development_account_operator_admin](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_iam_policy_document.cloudwatch_kms_merged](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_iam_policy_document.dynamodb_kms](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_iam_policy_document.dynamodb_kms_development_account_operator_admin](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_iam_policy_document.dynamodb_kms_merged](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_iam_policy_document.reduced_fees_uploads_s3_kms](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_iam_policy_document.reduced_fees_uploads_s3_kms_development_account_operator_admin](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_iam_policy_document.reduced_fees_uploads_s3_kms_merged](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_iam_policy_document.secrets_manager_kms](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_iam_policy_document.secrets_manager_kms_development_account_operator_admin](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_iam_policy_document.secrets_manager_kms_merged](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_iam_policy_document.sns_kms](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_iam_policy_document.sns_kms_development_account_operator_admin](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_iam_policy_document.sns_kms_merged](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_iam_policy_document.sqs_kms](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_iam_policy_document.sqs_kms_development_account_operator_admin](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_iam_policy_document.sqs_kms_merged](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_region.eu_west_1](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/region) | data source |
| [aws_region.eu_west_2](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/region) | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_accounts"></a> [accounts](#input\_accounts) | n/a | <pre>map(<br>    object({<br>      account_id    = string<br>      account_name  = string<br>      is_production = bool<br>      regions       = list(string)<br>    })<br>  )</pre> | n/a | yes |
| <a name="input_default_role"></a> [default\_role](#input\_default\_role) | n/a | `string` | `"modernising-lpa-ci"` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_resource_group_arns"></a> [resource\_group\_arns](#output\_resource\_group\_arns) | n/a |
| <a name="output_workspace_name"></a> [workspace\_name](#output\_workspace\_name) | n/a |
<!-- END_TF_DOCS -->
