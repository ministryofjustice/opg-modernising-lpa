data "aws_iam_policy_document" "lambda_access_policy" {
  statement {
    sid       = "allowLambdaInvoke"
    effect    = "Allow"
    resources = [module.schedule_runner.lambda.arn]
    actions = [
      "lambda:InvokeFunction",
    ]
  }
  provider = aws.region
}

resource "aws_iam_role_policy" "lambda_access_role_policy" {
  policy   = data.aws_iam_policy_document.lambda_access_policy.json
  role     = var.schedule_runner_scheduler.name
  provider = aws.region
}

data "aws_iam_policy_document" "schedule_runner" {
  statement {
    sid    = "${local.policy_region_prefix}DynamoDBEncryptionAccess"
    effect = "Allow"

    actions = [
      "kms:Encrypt",
      "kms:Decrypt",
      "kms:GenerateDataKey",
      "kms:RetireGrant",
    ]

    resources = [
      data.aws_kms_alias.dynamodb_encryption_key.target_key_arn,
    ]
  }

  statement {
    sid = "${local.policy_region_prefix}AllowDynamoDBAccess"

    actions = [
      "dynamodb:DeleteItem",
      "dynamodb:PutItem",
      "dynamodb:Query",
      "dynamodb:GetItem",
      "dynamodb:UpdateItem",
    ]

    resources = [
      var.lpas_table.arn,
      "${var.lpas_table.arn}/index/*",
    ]
  }

  statement {
    sid    = "${local.policy_region_prefix}AllowSecretAccess"
    effect = "Allow"

    actions = [
      "secretsmanager:GetSecretValue",
      "secretsmanager:DescribeSecret",
    ]

    resources = [
      data.aws_secretsmanager_secret.gov_uk_notify_api_key.arn,
      data.aws_secretsmanager_secret.lpa_store_jwt_key.arn,
      data.aws_secretsmanager_secret.lpa_store_jwt_secret_key.arn,
    ]
  }

  statement {
    effect = "Allow"

    resources = [
      data.aws_kms_alias.secrets_manager_secret_encryption_key.target_key_arn,
      data.aws_kms_alias.aws_lambda.target_key_arn,
    ]

    actions = [
      "kms:Decrypt",
      "kms:GenerateDataKey",
      "kms:GenerateDataKeyPair",
      "kms:GenerateDataKeyPairWithoutPlaintext",
      "kms:GenerateDataKeyWithoutPlaintext",
      "kms:DescribeKey",
    ]
  }

  statement {
    sid    = "${local.policy_region_prefix}AllowEventBusAccess"
    effect = "Allow"
    actions = [
      "events:PutEvents",
    ]
    resources = [
      var.event_bus.arn
    ]
  }

  statement {
    sid    = "${local.policy_region_prefix}XrayAccess"
    effect = "Allow"

    actions = [
      "xray:PutTraceSegments",
      "xray:PutTelemetryRecords",
      "xray:GetSamplingRules",
      "xray:GetSamplingTargets",
      "xray:GetSamplingStatisticSummaries",
    ]

    resources = ["*"]
  }

  provider = aws.region
}

resource "aws_iam_role_policy" "schedule_runner" {
  name     = "schedule_runner-${data.aws_default_tags.current.tags.environment-name}"
  role     = var.schedule_runner_lambda_role.id
  policy   = data.aws_iam_policy_document.schedule_runner.json
  provider = aws.region
}
