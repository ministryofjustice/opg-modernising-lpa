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

data "aws_network_acls" "default" {
  vpc_id = data.aws_vpc.default.id
  filter {
    name   = "default"
    values = ["true"]
  }
  provider = aws.region
}

resource "aws_default_network_acl" "default" {
  provider               = aws.region
  default_network_acl_id = data.aws_network_acls.default.ids[0]
  subnet_ids             = [for subnet in aws_default_subnet.default : subnet.id]

  egress {
    action     = "allow"
    cidr_block = "0.0.0.0/0"
    from_port  = 0
    icmp_code  = 0
    icmp_type  = 0
    protocol   = "-1" #tfsec:ignore:aws-ec2-no-excessive-port-access
    rule_no    = 100
    to_port    = 0
  }

  ingress {
    action     = "allow"
    cidr_block = "0.0.0.0/0"
    from_port  = 0
    icmp_code  = 0
    icmp_type  = 0
    protocol   = "-1" #tfsec:ignore:aws-ec2-no-excessive-port-access
    rule_no    = 100
    to_port    = 0
  }
  ingress {
    action     = "deny"
    cidr_block = "0.0.0.0/0" #tfsec:ignore:aws-ec2-no-public-ingress-acl
    from_port  = 22
    icmp_code  = 0
    icmp_type  = 0
    protocol   = "6"
    rule_no    = 120
    to_port    = 22
  }
  ingress {
    action     = "deny"
    cidr_block = "0.0.0.0/0" #tfsec:ignore:aws-ec2-no-public-ingress-acl
    from_port  = 3389
    icmp_code  = 0
    icmp_type  = 0
    protocol   = "6"
    rule_no    = 130
    to_port    = 3389
  }

}
