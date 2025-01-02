output "workspace_name" {
  value = terraform.workspace
}

variable "container_version" {
  type    = string
  default = "latest"
}

variable "public_access_enabled" {
  type    = bool
  default = false
}

variable "pagerduty_api_key" {
  type      = string
  sensitive = true
}

output "container_version" {
  value = var.container_version
}

output "public_access_enabled" {
  value = var.public_access_enabled
}

variable "environments" {
  type = map(
    object({
      account_id    = string
      account_name  = string
      is_production = bool
      regions       = list(string)
      app = object({
        env = object({
          app_public_url         = string
          auth_redirect_base_url = string
          notify_is_production   = string
          onelogin_url           = string
          dev_mode               = string
        })
        autoscaling = object({
          minimum = number
          maximum = number
        })
        dependency_health_check_alarm_enabled   = bool
        service_health_check_alarm_enabled      = bool
        cloudwatch_application_insights_enabled = bool
        fault_injection_experiments_enabled     = bool
        real_user_monitoring_cw_logs_enabled    = bool
      })
      mock_onelogin_enabled  = bool
      mock_pay_enabled       = bool
      egress_checker_enabled = bool
      uid_service = object({
        base_url = string
        api_arns = list(string)
      })
      lpa_store_service = object({
        base_url = string
        api_arns = list(string)
      })
      backups = object({
        backup_plan_enabled = bool
        copy_action_enabled = bool
      })
      dynamodb = object({
        table_name             = string
        region_replica_enabled = bool
        cloudtrail_enabled     = bool
      })
      ecs = object({
        fargate_spot_capacity_provider_enabled = bool
      })
      cloudwatch_log_groups = object({
        application_log_retention_days = number
      })
      application_load_balancer = object({
        deletion_protection_enabled = bool
        waf_alb_association_enabled = bool
      })
      cloudwatch_application_insights_enabled = bool
      pagerduty_service_name                  = string
      event_bus = object({
        target_event_bus_arn = string
        receive_account_ids  = list(string)
      })
      reduced_fees = object({
        enabled                                   = bool
        s3_object_replication_enabled             = bool
        target_environment                        = string
        destination_account_id                    = string
        enable_s3_batch_job_replication_scheduler = bool
      })
      s3_antivirus_provisioned_concurrency = number
    })
  )
}

locals {
  environment_name = lower(replace(terraform.workspace, "_", "-"))
  environment      = contains(keys(var.environments), local.environment_name) ? var.environments[local.environment_name] : var.environments["default"]

  mandatory_moj_tags = {
    business-unit    = "OPG"
    application      = "opg-modernising-lpa"
    environment-name = local.environment_name
    owner            = "OPG Webops: opgteam+modernising-lpa@digital.justice.gov.uk"
    is-production    = local.environment.is_production
    runbook          = "https://github.com/ministryofjustice/opg-modernising-lpa"
    source-code      = "https://github.com/ministryofjustice/opg-modernising-lpa"
  }

  optional_tags = {
    infrastructure-support = "OPG Webops: opgteam+modernising-lpa@digital.justice.gov.uk"
    account-name           = local.environment.account_name
  }

  default_tags = merge(local.mandatory_moj_tags, local.optional_tags)

  ecs_capacity_provider = local.environment.ecs.fargate_spot_capacity_provider_enabled ? "FARGATE_SPOT" : "FARGATE"

  mock_onelogin_version = "latest"

  search_index_name = "lpas_v2_${local.environment_name}"
}
