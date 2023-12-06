output "load_balancer" {
  value = aws_lb.app
}

output "load_balancer_security_group" {
  value = aws_security_group.app_loadbalancer
}

output "ecs_service" {
  value = aws_ecs_service.app
}

output "ecs_service_security_group_id" {
  value = aws_security_group.app_ecs_service.id
}
