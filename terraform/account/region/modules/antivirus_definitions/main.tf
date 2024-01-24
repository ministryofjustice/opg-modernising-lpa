resource "aws_cloudwatch_log_group" "lambda" {
  name              = "/aws/lambda/antivirus-definition-${data.aws_default_tags.current.tags.account-name}"
  kms_key_id        = aws_kms_key.cloudwatch.arn
  retention_in_days = 7
  provider          = aws.region
}

data "aws_security_group" "lambda_egress" {
  name     = "lambda-egress-${data.aws_region.current.name}"
  provider = aws.region
}

resource "aws_lambda_function" "lambda_function" {
  function_name = "antivirus-definition-${data.aws_default_tags.current.tags.account-name}"
  image_uri     = var.ecr_image_uri
  package_type  = "Image"
  role          = aws_iam_role.s3_antivirus_update.arn
  timeout       = 300
  memory_size   = 2048

  tracing_config {
    mode = "Active"
  }

  vpc_config {
    subnet_ids = data.aws_subnet.application[*].id
    security_group_ids = [
      data.aws_security_group.lambda_egress.id
    ]
  }

  environment {
    variables = {
      ANTIVIRUS_DEFINITIONS_BUCKET = aws_s3_bucket.bucket.id
    }
  }
  provider = aws.region
}

resource "aws_cloudwatch_event_rule" "cron" {
  name                = "antivirus-definition-cron"
  description         = "Updates the antivirus definitions"
  schedule_expression = "cron(00 03 * * ? *)"
  provider            = aws.region
}

resource "aws_cloudwatch_event_target" "lambda" {
  target_id = "runLambda"
  rule      = aws_cloudwatch_event_rule.cron.name
  arn       = aws_lambda_function.lambda_function.arn
  input     = "{}"
  provider  = aws.region
}

resource "aws_lambda_permission" "cloudwatch" {
  statement_id  = "AllowExecutionFromCloudWatch"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.lambda_function.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.cron.arn
  provider      = aws.region
}
