output "app_load_balancer" {
  value = module.app.load_balancer
}

output "app_load_balancer_security_group" {
  value = module.app.load_balancer_security_group
}

output "vpc" {
  value = data.aws_vpc.main
}
