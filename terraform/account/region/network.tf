module "network" {
  source                         = "github.com/ministryofjustice/opg-terraform-aws-network?ref=v1.1.0"
  cidr                           = var.network_cidr_block
  default_security_group_ingress = []
  default_security_group_egress  = []
  providers = {
    aws = aws.region
  }
}

data "aws_vpc" "default" {
  provider = aws.region
  default  = true
}

resource "aws_default_security_group" "default" {
  provider = aws.region
  vpc_id   = data.aws_vpc.default.id
  ingress  = []
  egress   = []
}

data "aws_availability_zones" "available" {
  state    = "available"
  provider = aws.region
}

resource "aws_default_subnet" "default" {
  for_each                = toset(data.aws_availability_zones.available.names)
  availability_zone       = each.value
  map_public_ip_on_launch = false
  provider                = aws.region
}
