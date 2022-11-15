data "aws_caller_identity" "eu_west_1" {
  provider = aws.eu_west_1
}

data "aws_region" "eu_west_1" {
  provider = aws.eu_west_1
}

data "aws_iam_role" "rum_monitor_unauthenticated" {
  name     = "RUM-Monitor-${data.aws_region.eu_west_1.name}"
  provider = aws.eu_west_1
}

# create this policy and attachment for each environment
resource "aws_iam_role_policy" "rum_monitor_unauthenticated" {
  count    = local.environment.app.rum_enabled ? 1 : 0
  name     = "RUMPutBatchMetrics-${local.environment_name}"
  policy   = data.aws_iam_policy_document.rum_monitor_unauthenticated.json
  role     = data.aws_iam_role.rum_monitor_unauthenticated.id
  provider = aws.eu_west_1
}

data "aws_iam_policy_document" "rum_monitor_unauthenticated" {
  statement {
    effect = "Allow"
    resources = [
      "arn:aws:rum:${data.aws_region.eu_west_1.name}:${data.aws_caller_identity.eu_west_1.account_id}:appmonitor/${local.environment_name}"
    ]
    actions = [
      "rum:PutRumEvents",
    ]
  }
  provider = aws.eu_west_1
}

data "aws_ssm_parameter" "rum_monitor_identity_pool_id" {
  name     = "rum_monitor_identity_pool_id"
  provider = aws.eu_west_1
}

resource "aws_rum_app_monitor" "main" {
  name           = local.environment_name
  domain         = "*.${aws_route53_record.app.fqdn}"
  cw_log_enabled = true
  app_monitor_configuration {
    allow_cookies       = true
    enable_xray         = true
    identity_pool_id    = data.aws_ssm_parameter.rum_monitor_identity_pool_id.value
    session_sample_rate = 1.0
    telemetries = [
      "errors",
      "http",
      "performance",
    ]
  }
  provider = aws.eu_west_1
}
