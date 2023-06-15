resource "aws_vpc_endpoint" "execute_api" {
  vpc_id            = data.aws_vpc.main.id
  service_name      = "com.amazonaws.eu-west-1.execute-api"
  vpc_endpoint_type = "Interface"
  subnet_ids        = data.aws_subnet.application.*.id

  security_group_ids = [
    aws_security_group.execute_api.id,
  ]

  private_dns_enabled = true
  provider            = aws.region
}

resource "aws_vpc_endpoint_policy" "app_ecs_access" {
  vpc_endpoint_id = aws_vpc_endpoint.execute_api.id
  policy          = data.aws_iam_policy_document.execute_api.json
  provider        = aws.region
}

data "aws_iam_policy_document" "execute_api" {
  statement {
    effect = "Allow"
    sid    = "AllowAll"
    actions = [
      "execute-api:Invoke",
    ]
    principals {
      type = "AWS"
      identifiers = [
        var.ecs_task_roles.app.arn,
      ]
    }
  }
}

resource "aws_security_group" "execute_api" {
  name        = "execute-api"
  description = "For execute-api vpc endpoint"
  vpc_id      = data.aws_vpc.main.id
  provider    = aws.region
}

# --------

data "aws_security_group" "execute_api" {
  name     = aws_security_group.execute_api.name
  vpc_id   = data.aws_vpc.main.id
  provider = aws.region
}

resource "aws_vpc_security_group_ingress_rule" "example" {
  security_group_id            = data.aws_security_group.execute_api.id
  from_port                    = 443
  to_port                      = 443
  ip_protocol                  = "tcp"
  referenced_security_group_id = module.app.app_ecs_service_security_group.id
  provider                     = aws.region
}

resource "aws_ssm_parameter" "execute_api_id" {
  name     = "/modernising-lpa/execute-api-id${data.aws_default_tags.current.tags.environment-name}"
  type     = "String"
  value    = aws_vpc_endpoint.execute_api.id
  provider = aws.management_global
}
