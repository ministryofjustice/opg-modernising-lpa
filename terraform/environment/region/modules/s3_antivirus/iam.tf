resource "aws_iam_role_policy" "lambda" {
  name     = "s3-antivirus-policy"
  role     = var.lambda_task_role.id
  policy   = data.aws_iam_policy_document.lambda.json
  provider = aws.region
}

data "aws_iam_policy_document" "lambda" {
  statement {
    sid       = "allowLogging"
    effect    = "Allow"
    resources = [aws_cloudwatch_log_group.lambda.arn]
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
      "logs:DescribeLogStreams"
    ]
  }

  statement {
    sid       = "allowS3Tagging"
    effect    = "Allow"
    resources = [var.data_store_bucket.arn, "${var.data_store_bucket.arn}/*"]
    actions = [
      "s3:GetBucketLocation",
      "s3:GetObject",
      "s3:GetObjectTagging",
      "s3:PutObjectTagging"
    ]
  }

  statement {
    sid       = "allowS3ObjectDecryption"
    effect    = "Allow"
    resources = [data.aws_kms_alias.uploads_encryption_key.target_key_arn]
    actions = [
      "kms:Decrypt",
      "kms:DescribeKey",
      "kms:RetireGrant",
    ]
  }

  statement {
    sid       = "allowVirusDefinitions"
    effect    = "Allow"
    resources = [var.definition_bucket.arn, "${var.definition_bucket.arn}/*"]
    actions = [
      "s3:GetBucketLocation",
      "s3:GetObject"
    ]
  }

  statement {
    sid       = "tracing"
    effect    = "Allow"
    resources = ["*"]
    actions = [
      "xray:PutTraceSegments",
      "xray:PutTelemetryRecords",
      "xray:GetSamplingRules",
      "xray:GetSamplingTargets",
      "xray:GetSamplingStatisticSummaries"
    ]
  }
  provider = aws.region
}

resource "aws_lambda_permission" "allow_lambda_execution_operator" {
  statement_id  = "AllowExecutionOperator"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.lambda_function.function_name
  principal     = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/operator"
  provider      = aws.region
}
