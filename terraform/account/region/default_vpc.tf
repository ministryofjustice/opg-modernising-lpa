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
  ingress {
    action          = "deny"
    ipv6_cidr_block = "::/0"
    from_port       = 22
    icmp_code       = 0
    icmp_type       = 0
    protocol        = "6"
    rule_no         = 125
    to_port         = 22
  }
  ingress {
    action          = "deny"
    ipv6_cidr_block = "::/0"
    from_port       = 3389
    icmp_code       = 0
    icmp_type       = 0
    protocol        = "6"
    rule_no         = 135
    to_port         = 3389
  }
  ingress {
    action     = "allow"
    cidr_block = "0.0.0.0/0"
    from_port  = 0
    icmp_code  = 0
    icmp_type  = 0
    protocol   = "-1" #tfsec:ignore:aws-ec2-no-excessive-port-access
    rule_no    = 160
    to_port    = 0
  }
}

resource "aws_flow_log" "default_vpc" {
  provider                 = aws.region
  log_destination_type     = "cloud-watch-logs"
  log_destination          = aws_cloudwatch_log_group.default_vpc_flow_log.arn
  log_format               = null
  iam_role_arn             = aws_iam_role.default_vpc_flow_log_cloudwatch.arn
  traffic_type             = "ALL"
  vpc_id                   = data.aws_vpc.default.id
  max_aggregation_interval = 600
}

resource "aws_cloudwatch_log_group" "default_vpc_flow_log" {
  provider          = aws.region
  name              = "/aws/vpc-flow-log/${data.aws_vpc.default.id}"
  retention_in_days = 400
  kms_key_id        = data.aws_kms_alias.cloudwatch_application_logs_encryption.target_key_arn
}

resource "aws_iam_role" "default_vpc_flow_log_cloudwatch" {
  provider           = aws.region
  name_prefix        = "default-vpc-flow-log-role-"
  assume_role_policy = data.aws_iam_policy_document.default_vpc_flow_log_cloudwatch_assume_role.json
}

data "aws_iam_policy_document" "default_vpc_flow_log_cloudwatch_assume_role" {
  provider = aws.region
  statement {
    principals {
      type        = "Service"
      identifiers = ["vpc-flow-logs.amazonaws.com"]
    }

    effect = "Allow"

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role_policy_attachment" "default_vpc_flow_log_cloudwatch" {
  provider   = aws.region
  role       = aws_iam_role.default_vpc_flow_log_cloudwatch.name
  policy_arn = aws_iam_policy.default_vpc_flow_log_cloudwatch.arn
}

resource "aws_iam_policy" "default_vpc_flow_log_cloudwatch" {
  provider    = aws.region
  name_prefix = "vpc-flow-log-to-cloudwatch-"
  policy      = data.aws_iam_policy_document.default_vpc_flow_log_cloudwatch.json
}

#tfsec:ignore:aws-iam-no-policy-wildcards
data "aws_iam_policy_document" "default_vpc_flow_log_cloudwatch" {
  provider = aws.region
  statement {
    sid = "AWSDefaultVPCFlowLogsPushToCloudWatch"

    effect = "Allow"

    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
      "logs:DescribeLogGroups",
      "logs:DescribeLogStreams",
    ]

    resources = ["*"]
  }
}


#  Default VPC Endpoint
resource "aws_security_group" "default_vpc_endpoints" {
  provider    = aws.region
  name        = "default-vpc-endpoint-access-private-subnets-${data.aws_region.current.name}"
  description = "Default VPC Interface Endpoints Security Group"
  vpc_id      = data.aws_vpc.default.id
  tags        = { Name = "default-vpc-endpoint-access-private-subnets-${data.aws_region.current.name}" }
}

resource "aws_security_group_rule" "default_vpc_endpoints_subnet_ingress" {
  provider          = aws.region
  from_port         = 443
  to_port           = 443
  protocol          = "tcp"
  security_group_id = aws_security_group.default_vpc_endpoints.id
  type              = "ingress"
  cidr_blocks       = [for subnet in aws_default_subnet.default : subnet.cidr_block]
  description       = "Allow Services in Private Subnets of ${data.aws_region.current.name} to connect to Default VPC Interface Endpoints"
}

locals {
  default_vpc_interface_endpoint = toset([
    "ec2",
  ])
}

resource "aws_vpc_endpoint" "default_private" {
  provider = aws.region
  for_each = local.default_vpc_interface_endpoint

  vpc_id              = data.aws_vpc.default.id
  service_name        = "com.amazonaws.${data.aws_region.current.name}.${each.value}"
  vpc_endpoint_type   = "Interface"
  private_dns_enabled = true
  security_group_ids  = aws_security_group.default_vpc_endpoints[*].id
  subnet_ids          = [for subnet in aws_default_subnet.default : subnet.id]
  tags                = { Name = "default-vpc-${each.value}-private-${data.aws_region.current.name}" }
}

resource "aws_vpc_endpoint_policy" "default_vpc_ec2" {
  provider        = aws.region
  vpc_endpoint_id = aws_vpc_endpoint.default_private["ec2"].id
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
