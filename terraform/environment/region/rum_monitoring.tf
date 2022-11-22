data "aws_iam_role" "rum_monitor_unauthenticated" {
  name     = "RUM-Monitor-Unauthenticated-${data.aws_region.current.name}"
  provider = aws.global
}

# create this policy and attachment for each environment
resource "aws_iam_role_policy" "rum_monitor_unauthenticated" {
  name     = "RUMPutBatchMetrics-${data.aws_default_tags.current.tags.environment-name}"
  policy   = data.aws_iam_policy_document.rum_monitor_unauthenticated.json
  role     = data.aws_iam_role.rum_monitor_unauthenticated.id
  provider = aws.global
}

data "aws_iam_policy_document" "rum_monitor_unauthenticated" {
  statement {
    effect = "Allow"
    resources = [
      "arn:aws:rum:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:appmonitor/${data.aws_default_tags.current.tags.environment-name}",
    ]
    actions = [
      "rum:PutRumEvents",
    ]
  }
  provider = aws.global
}

data "aws_secretsmanager_secret_version" "rum_monitor_identity_pool_id" {
  secret_id = var.rum_monitor_identity_pool_id_secretsmanager_secret_id
  provider  = aws.region
}

locals {
  dns_namespace_for_environment = data.aws_default_tags.current.tags.environment-name == "production" ? "" : "${data.aws_default_tags.current.tags.environment-name}."
}

resource "aws_rum_app_monitor" "main" {
  name           = data.aws_default_tags.current.tags.environment-name
  domain         = "${local.dns_namespace_for_environment}app.modernising.opg.service.justice.gov.uk"
  cw_log_enabled = true
  app_monitor_configuration {
    allow_cookies       = true
    enable_xray         = true
    identity_pool_id    = data.aws_secretsmanager_secret_version.rum_monitor_identity_pool_id.secret_string
    session_sample_rate = 1.0
    telemetries = [
      "errors",
      "http",
      "performance",
    ]
  }
  provider = aws.region
}

resource "aws_secretsmanager_secret_version" "rum_monitor_application_id" {
  secret_id     = var.rum_monitor_application_id_secretsmanager_secret_id
  secret_string = aws_rum_app_monitor.main.app_monitor_id
  provider      = aws.region
}
