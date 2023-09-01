resource "aws_iam_role" "s3_antivirus_update" {
  name               = "s3-antivirus-update-${data.aws_default_tags.current.tags.account-name}"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume.json
  provider           = aws.region
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

resource "aws_iam_role_policy_attachment" "vpc_execution_role" {
  role       = aws_iam_role.s3_antivirus_update.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"
  provider   = aws.region
}

resource "aws_iam_role_policy" "lambda" {
  name     = "s3-antivirus-policy"
  role     = aws_iam_role.s3_antivirus_update.id
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
    sid       = "allowS3LocateList"
    effect    = "Allow"
    resources = [aws_s3_bucket.bucket.arn]
    actions = [
      "s3:GetBucketLocation",
      "s3:ListBucket",
    ]
  }

  #tfsec:ignore:aws-iam-no-policy-wildcards
  statement {
    sid       = "allowS3GetPut"
    effect    = "Allow"
    resources = [aws_s3_bucket.bucket.arn, "${aws_s3_bucket.bucket.arn}/*"]
    actions = [
      "s3:GetObject",
      "s3:PutObject"
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
