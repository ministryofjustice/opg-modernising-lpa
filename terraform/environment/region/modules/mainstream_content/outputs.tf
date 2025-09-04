output "load_balancer" {
  value = aws_lb.app
}

output "load_balancer_security_group" {
  value = aws_security_group.app_loadbalancer
}
