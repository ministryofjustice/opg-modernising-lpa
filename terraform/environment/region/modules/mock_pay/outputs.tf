output "load_balancer" {
  value = aws_lb.mock_pay
}

output "load_balancer_security_group" {
  value = aws_security_group.mock_pay_loadbalancer
}

output "ecs_service" {
  value = aws_ecs_service.mock_pay
}
