output "resource_group_arns" {
  value = [
    module.eu_west_1[0].resource_group_arn,
    contains(local.environment.regions, "eu-west-2") ? module.eu_west_2[0].resource_group_arn : null,
    module.global.resource_group_arn,
  ]
}

output "app_fqdn" {
  value = contains(local.environment.regions, "eu-west-1") ? module.eu_west_1[0].app_fqdn : module.eu_west_2[0].app_fqdn
}

locals {
  environment_config = {
    region                              = "eu-west-1"
    account_id                          = local.environment.account_id
    app_load_balancer_security_group_id = module.eu_west_1[0].app_load_balancer_security_group.id
  }
}

output "environment_config_json" {
  value = jsonencode(local.environment_config)
}
