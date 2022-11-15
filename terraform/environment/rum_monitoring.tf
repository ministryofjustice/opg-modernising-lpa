data "aws_caller_identity" "global" {
  provider = aws.global
}

data "aws_region" "global" {
  provider = aws.global
}

data "aws_iam_role" "rum_monitor_unauthenticated" {
  count    = local.environment.app.rum_enabled ? 1 : 0
  name     = "RUM-Monitor-Unauthenticated"
  provider = aws.global
}

# create this policy and attachment for each environment
resource "aws_iam_role_policy" "rum_monitor_unauthenticated" {
  count    = local.environment.app.rum_enabled ? 1 : 0
  name     = "RUMPutBatchMetrics-${local.environment_name}"
  policy   = data.aws_iam_policy_document.rum_monitor_unauthenticated[0].json
  role     = data.aws_iam_role.rum_monitor_unauthenticated[0].id
  provider = aws.global
}

data "aws_iam_policy_document" "rum_monitor_unauthenticated" {
  count = local.environment.app.rum_enabled ? 1 : 0
  statement {
    effect = "Allow"
    resources = [
      "arn:aws:rum:eu-west-1:${data.aws_caller_identity.global.account_id}:appmonitor/${local.environment_name}",
      "arn:aws:rum:eu-west-2:${data.aws_caller_identity.global.account_id}:appmonitor/${local.environment_name}"
    ]
    actions = [
      "rum:PutRumEvents",
    ]
  }
  provider = aws.global
}

data "aws_ssm_parameter" "rum_monitor_identity_pool_id" {
  count    = local.environment.app.rum_enabled ? 1 : 0
  name     = "rum_monitor_identity_pool_id"
  provider = aws.eu_west_1
}

resource "aws_rum_app_monitor" "main" {
  count          = local.environment.app.rum_enabled ? 1 : 0
  name           = local.environment_name
  domain         = "*.${aws_route53_record.app.fqdn}"
  cw_log_enabled = true
  app_monitor_configuration {
    allow_cookies       = true
    enable_xray         = true
    identity_pool_id    = data.aws_ssm_parameter.rum_monitor_identity_pool_id[0].value
    session_sample_rate = 1.0
    telemetries = [
      "errors",
      "http",
      "performance",
    ]
  }
  provider = aws.eu_west_1
}
