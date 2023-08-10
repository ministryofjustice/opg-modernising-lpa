resource "aws_iam_role" "lambda_role" {
  name               = "${var.lambda_name}-${var.environment}"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume.json
  lifecycle {
    create_before_destroy = true
  }
  provider = aws.region
}

data "aws_iam_policy_document" "lambda_assume" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
  provider = aws.region
}

data "aws_iam_policy" "aws_xray_write_only_access" {
  arn      = "arn:aws:iam::aws:policy/AWSXrayWriteOnlyAccess"
  provider = aws.region
}

resource "aws_iam_role_policy_attachment" "aws_xray_write_only_access" {
  role       = aws_iam_role.lambda_role.name
  policy_arn = data.aws_iam_policy.aws_xray_write_only_access.arn
  provider   = aws.region
}

resource "aws_iam_role_policy" "lambda" {
  name     = "lambda-${var.environment}"
  role     = aws_iam_role.lambda_role.id
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

  provider = aws.region
}

resource "aws_iam_role_policy_attachment" "vpc_access_execution_role" {
  role       = aws_iam_role.lambda_role.name
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
