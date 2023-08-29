module "event_received" {
  source      = "../lambda"
  lambda_name = "event-received"
  description = "Function to react when an event is recieved"
  environment_variables = {
    LPAS_TABLE = var.lpas_table.name
  }
  image_uri   = "${var.lambda_function_image_ecr_url}:${var.lambda_function_image_tag}"
  ecr_arn     = var.lambda_function_image_ecr_arn
  environment = data.aws_default_tags.current.tags.environment-name
  kms_key     = data.aws_kms_alias.cloudwatch_application_logs_encryption.target_key_arn
  timeout     = 300
  memory      = 1024
  providers = {
    aws.region = aws.region
  }
}

resource "aws_cloudwatch_event_rule" "receive_events" {
  name           = "${data.aws_default_tags.current.tags.environment-name}-receive-events"
  description    = "receive events from sirius"
  event_bus_name = var.event_bus_name

  event_pattern = jsonencode({
    source      = ["opg.poas.sirius"],
    detail-type = ["evidence-received", "fee-approved"]
  })
  provider = aws.region
}

resource "aws_cloudwatch_event_target" "receive_events" {
  target_id      = "${data.aws_default_tags.current.tags.environment-name}-receive-events"
  event_bus_name = var.event_bus_name
  rule           = aws_cloudwatch_event_rule.receive_events.name
  arn            = module.event_received.lambda.arn
  provider       = aws.region
}

resource "aws_lambda_permission" "allow_cloudwatch_to_call_event_received" {
  statement_id   = "AllowExecutionFromCloudWatch"
  action         = "lambda:InvokeFunction"
  function_name  = module.event_received.lambda.function_name
  principal      = "events.amazonaws.com"
  source_account = data.aws_caller_identity.current.account_id
  source_arn     = aws_cloudwatch_event_rule.receive_events.arn
  provider       = aws.region
}

resource "aws_iam_role_policy" "event_received" {
  name     = "event_received-${data.aws_default_tags.current.tags.environment-name}"
  role     = module.event_received.lambda_role.id
  policy   = data.aws_iam_policy_document.event_received.json
  provider = aws.region
}

data "aws_kms_alias" "dynamodb_encryption_key" {
  name     = "alias/${data.aws_default_tags.current.tags.application}_dynamodb_encryption"
  provider = aws.region
}

locals {
  policy_region_prefix = lower(replace(data.aws_region.current.name, "-", ""))
}

data "aws_iam_policy_document" "event_received" {
  statement {
    sid    = "${local.policy_region_prefix}DynamoDBEncryptionAccess"
    effect = "Allow"

    actions = [
      "kms:Encrypt",
      "kms:Decrypt",
      "kms:GenerateDataKey",
    ]

    resources = [
      data.aws_kms_alias.dynamodb_encryption_key.target_key_arn,
    ]
  }

  statement {
    sid = "${local.policy_region_prefix}Allow"

    actions = [
      "dynamodb:PutItem",
      "dynamodb:Query",
    ]

    resources = [
      var.lpas_table.arn,
      "${var.lpas_table.arn}/index/*",
    ]
  }

  statement {
    sid    = "${local.policy_region_prefix}SecretAccess"
    effect = "Allow"

    actions = [
      "secretsmanager:GetSecretValue",
      "secretsmanager:DescribeSecret",
    ]

    resources = [
      data.aws_secretsmanager_secret.gov_uk_notify_api_key.arn,
    ]
  }

  provider = aws.region
}
