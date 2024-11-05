resource "aws_security_group" "vpc_endpoints_private" {
  provider    = aws.region
  name        = "vpc-endpoint-access-private-subnets"
  description = "VPC Interface Endpoints Security Group"
  vpc_id      = module.network.vpc.id
  tags        = { Name = "vpc-endpoint-access-private-subnets" }
}

resource "aws_security_group_rule" "vpc_endpoints_private_subnet_ingress" {
  provider          = aws.region
  from_port         = 443
  to_port           = 443
  protocol          = "tcp"
  security_group_id = aws_security_group.vpc_endpoints_private.id
  type              = "ingress"
  cidr_blocks       = module.network.application_subnets[*].cidr_block
  description       = "Allow Services in Private Subnets of ${data.aws_region.current.name} to connect to VPC Interface Endpoints"
}

resource "aws_security_group_rule" "vpc_endpoints_public_subnet_ingress" {
  provider          = aws.region
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
    "ecr.api",
    "ecr.dkr",
    "execute-api",
    "events",
    "logs",
    "rum",
    "secretsmanager",
    "ssm",
    "xray",
  ])
}

resource "aws_vpc_endpoint" "private" {
  provider = aws.region
  for_each = local.interface_endpoint

  vpc_id              = module.network.vpc.id
  service_name        = "com.amazonaws.${data.aws_region.current.name}.${each.value}"
  vpc_endpoint_type   = "Interface"
  private_dns_enabled = true
  security_group_ids  = aws_security_group.vpc_endpoints_private[*].id
  subnet_ids          = module.network.application_subnets[*].id
  tags                = { Name = "${each.value}-private" }
}

resource "aws_vpc_endpoint_policy" "private" {
  provider        = aws.region
  for_each        = local.interface_endpoint
  vpc_endpoint_id = aws_vpc_endpoint.private[each.value].id
  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Sid" : "AllowAll",
        "Effect" : "Allow",
        "Principal" : {
          "AWS" : "arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"
        },
        "Action" : [
          "${startswith(each.value, "ecr") ? "ecr" : each.value}:*"
        ],
        "Resource" : "*"
      }
    ]
  })
}

data "aws_route_tables" "application" {
  provider = aws.region
  filter {
    name   = "tag:Name"
    values = ["application-route-table"]
  }
}

resource "aws_vpc_endpoint" "s3" {
  provider          = aws.region
  vpc_id            = module.network.vpc.id
  service_name      = "com.amazonaws.${data.aws_region.current.name}.s3"
  route_table_ids   = tolist(data.aws_route_tables.application.ids)
  vpc_endpoint_type = "Gateway"
  policy            = data.aws_iam_policy_document.s3.json
  tags              = { Name = "s3-private" }
}

resource "aws_vpc_endpoint" "dynamodb" {
  provider          = aws.region
  vpc_id            = module.network.vpc.id
  service_name      = "com.amazonaws.${data.aws_region.current.name}.dynamodb"
  route_table_ids   = tolist(data.aws_route_tables.application.ids)
  vpc_endpoint_type = "Gateway"
  policy            = data.aws_iam_policy_document.allow_account_access.json
  tags              = { Name = "dynamodb-private" }
}



data "aws_iam_policy_document" "allow_account_access" {
  provider = aws.region
  statement {
    sid       = "Allow-callers-from-specific-account"
    effect    = "Allow"
    actions   = ["*"]
    resources = ["*"]
    principals {
      type        = "AWS"
      identifiers = ["*"]
    }
    condition {
      test     = "StringEquals"
      variable = "aws:PrincipalAccount"
      values   = [data.aws_caller_identity.current.account_id]
    }
  }
}

data "aws_iam_policy_document" "s3" {
  source_policy_documents = [
    data.aws_iam_policy_document.allow_account_access.json,
    data.aws_iam_policy_document.s3_bucket_access.json,
  ]
}

data "aws_iam_policy_document" "s3_bucket_access" {
  statement {
    sid       = "Access-to-specific-bucket-only"
    effect    = "Allow"
    actions   = ["s3:GetObject"]
    resources = ["arn:aws:s3:::prod-${data.aws_region.current.name}-starport-layer-bucket/*"]
    principals {
      type        = "AWS"
      identifiers = ["*"]
    }
  }
}

resource "aws_opensearchserverless_vpc_endpoint" "lpas_collection_vpc_endpoint" {
  name               = "opensearch"
  vpc_id             = module.network.vpc.id
  subnet_ids         = module.network.application_subnets[*].id
  security_group_ids = aws_security_group.vpc_endpoints_private[*].id
  provider           = aws.region
}
