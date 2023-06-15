resource "aws_vpc_endpoint" "execute_api" {
  vpc_id            = data.aws_vpc.main.id
  service_name      = "com.amazonaws.eu-west-1.execute-api"
  vpc_endpoint_type = "Interface"
  subnet_ids        = data.aws_subnet.application.*.id

  security_group_ids = [
    aws_security_group.execute_api.id,
  ]

  private_dns_enabled = true
}

resource "aws_security_group" "execute_api" {
  name        = "execute-api"
  description = "For execute-api vpc endpoint"
  vpc_id      = data.aws_vpc.main.id
}

# resource "aws_security_group_rule" "execute_api" {
#   type              = "ingress"
#   from_port         = 443
#   to_port           = 443
#   protocol          = "tcp"
#   source_security_group_id = ""
# }
