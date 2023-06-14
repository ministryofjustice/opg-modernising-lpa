output "app_load_balancer" {
  value = module.app.load_balancer
}

output "app_load_balancer_security_group" {
  value = module.app.load_balancer_security_group
}

output "resource_group_arn" {
  value = aws_resourcegroups_group.environment.arn
}

output "app_fqdn" {
  value = aws_route53_record.app.fqdn
}
