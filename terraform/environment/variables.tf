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
      app = object({
        env = object({
          app_public_url = string
        })
      })
      backups = object({
        backup_plan_enabled = bool
        copy_action_enabled = bool
      })
      ecs = object({
        enable_fargate_spot_capacity_provider = bool
      })
      cloudwatch_log_groups = object({
        application_log_retention_days = number
      })
      application_load_balancer = object({
        enable_deletion_protection = bool
      })
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
  }

  default_tags = merge(local.mandatory_moj_tags, local.optional_tags)

  ecs_capacity_provider = local.environment.ecs.enable_fargate_spot_capacity_provider ? "FARGATE_SPOT" : "FARGATE"
}
