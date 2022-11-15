data "aws_caller_identity" "eu_west_1" {
  provider = aws.eu_west_1
}

data "aws_region" "eu_west_1" {
  provider = aws.eu_west_1
}

data "aws_iam_role" "rum_monitor_unauthenticated" {
  name     = "RUM-Monitor-eu-west-1-653761790766-0155138158661-Unauth"
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
