data "aws_iam_policy" "aws_xray_write_only_access" {
  arn      = "arn:aws:iam::aws:policy/AWSXrayWriteOnlyAccess"
  provider = aws.region
}

resource "aws_iam_role_policy_attachment" "aws_xray_write_only_access" {
  role       = var.aws_iam_role.name
  policy_arn = data.aws_iam_policy.aws_xray_write_only_access.arn
  provider   = aws.region
}

resource "aws_iam_role_policy" "lambda" {
  name     = "lambda-${var.environment}"
  role     = var.aws_iam_role.id
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
    sid    = "logEncryption"
    effect = "Allow"
    resources = [
      data.aws_kms_alias.log_encryption_key.arn
    ]
    actions = [
      "kms:Encrypt",
      "kms:ReEncrypt*",
      "kms:GenerateDataKey*",
      "kms:DescribeKey"
    ]
  }

  override_policy_documents = var.iam_policy_documents
  provider                  = aws.region
}

resource "aws_iam_role_policy_attachment" "vpc_access_execution_role" {
  role       = var.aws_iam_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"
  provider   = aws.region
}

resource "aws_lambda_permission" "allow_lambda_execution_operator" {
  statement_id  = "AllowExecutionOperator"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.lambda_function.function_name
  principal     = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/operator"
  provider      = aws.region
}
