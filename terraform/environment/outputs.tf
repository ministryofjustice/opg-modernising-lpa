output "resource_group_arns" {
  value = [
    module.eu_west_1[0].resource_group_arns,
    contains(local.environment.regions, "eu-west-2") ? module.eu_west_2[0].resource_group_arns : null,
    module.global[0].resource_group_arns,
  ]
}
