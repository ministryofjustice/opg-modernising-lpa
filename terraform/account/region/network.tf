module "network" {
  source                         = "github.com/ministryofjustice/opg-terraform-aws-network?ref=v1.5.0"
  cidr                           = var.network_cidr_block
  enable_dns_hostnames           = true
  enable_dns_support             = true
  default_security_group_ingress = []
  default_security_group_egress  = []
  providers = {
    aws = aws.region
  }
}

resource "aws_security_group" "lambda_egress" {
  name        = "lambda-egress-${data.aws_region.current.name}"
  vpc_id      = module.network.vpc.id
  description = "Shared security group lambda for outbound traffic"

  tags     = { "Name" = "lambda-egress-${data.aws_region.current.name}" }
  provider = aws.region
}

resource "aws_security_group_rule" "lambda_egress" {
  type              = "egress"
  protocol          = "-1"
  from_port         = 0
  to_port           = 0
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.lambda_egress.id
  description       = "Outbound Lambda"
  provider          = aws.region
}
