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
  kms_key_id        = var.flow_log_cloudwatch_log_group_kms_key_id
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
