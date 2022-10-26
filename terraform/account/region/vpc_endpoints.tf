resource "aws_security_group" "vpc_endpoints_private" {
  name        = "vpc-endpoint-access-private-subnets-${data.aws_region.current.name}"
  description = "VPC Interface Endpoints Security Group"
  vpc_id      = module.network.vpc.id
  tags        = { Name = "vpc-endpoint-access-private-subnets-${data.aws_region.current.name}" }
}

resource "aws_security_group_rule" "vpc_endpoints_private_subnet_ingress" {
  from_port         = 443
  to_port           = 443
  protocol          = "tcp"
  security_group_id = aws_security_group.vpc_endpoints_private.id
  type              = "ingress"
  cidr_blocks       = module.network.application_subnets[*].cidr_block
  description       = "Allow Services in Private Subnets of ${data.aws_region.current.name} to connect to VPC Interface Endpoints"
}

resource "aws_security_group_rule" "vpc_endpoints_public_subnet_ingress" {
  from_port         = 443
  to_port           = 443
  protocol          = "tcp"
  security_group_id = aws_security_group.vpc_endpoints_private.id
  type              = "ingress"
  cidr_blocks       = module.network.public_subnets[*].cidr_block
  description       = "Allow Services in Public Subnets of ${data.aws_region.current.name} to connect to VPC Interface Endpoints"
}

locals {
  interface_endpoint = toset([
    "ec2",
  ])
}

resource "aws_vpc_endpoint" "private" {
  for_each = local.interface_endpoint

  vpc_id              = module.network.vpc.id
  service_name        = "com.amazonaws.${data.aws_region.current.name}.${each.value}"
  vpc_endpoint_type   = "Interface"
  private_dns_enabled = true
  security_group_ids  = aws_security_group.vpc_endpoints_private[*].id
  subnet_ids          = module.network.application_subnets[*].cidr_block
  tags                = { Name = "${each.value}-private-${data.aws_region.current.name}" }
}

resource "aws_vpc_endpoint_policy" "ec2" {
  vpc_endpoint_id = aws_vpc_endpoint.private["ec2"].id
  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Sid" : "AllowAll",
        "Effect" : "Allow",
        "Principal" : {
          "AWS" : "*"
        },
        "Action" : [
          "ec2:*"
        ],
        "Resource" : "*"
      }
    ]
  })
}
