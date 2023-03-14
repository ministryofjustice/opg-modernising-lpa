output "workspace_name" {
  value = terraform.workspace
}

variable "container_version" {
  type    = string
  default = "latest"
}

output "container_version" {
  value = var.container_version
}

variable "environments" {
  type = map(
    object({
      account_id    = string
      account_name  = string
      is_production = bool
      regions       = list(string)
      app = object({
        public_access_enabled = bool
        env = object({
          app_public_url         = string
          auth_redirect_base_url = string
          notify_is_production   = string
          yoti_client_sdk_id     = string
          yoti_scenario_id       = string
        })
      })
      backups = object({
        backup_plan_enabled = bool
        copy_action_enabled = bool
      })
      dynamodb = object({
        region_replica_enabled = bool
        stream_enabled         = bool
      })
      ecs = object({
        fargate_spot_capacity_provider_enabled = bool
      })
      cloudwatch_log_groups = object({
        application_log_retention_days = number
      })
      application_load_balancer = object({
        deletion_protection_enabled = bool
      })
      cloudwatch_application_insights_enabled = bool
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
}
