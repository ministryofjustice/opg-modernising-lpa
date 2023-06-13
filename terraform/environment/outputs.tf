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

output "environment_config_json" {
  value = jsonencode(local.environment_config)
}
