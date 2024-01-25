data "aws_kms_alias" "cloudwatch_application_logs_encryption" {
  name     = "alias/${data.aws_default_tags.current.tags.application}_cloudwatch_application_logs_encryption"
  provider = aws.region
}

resource "aws_cloudwatch_log_group" "lambda" {
  name              = "/aws/lambda/s3-antivirus-${data.aws_default_tags.current.tags.environment-name}"
  kms_key_id        = data.aws_kms_alias.cloudwatch_application_logs_encryption.target_key_arn
  retention_in_days = 30
  provider          = aws.region
}

resource "aws_lambda_function" "lambda_function" {
  function_name = "s3-antivirus-${data.aws_default_tags.current.tags.environment-name}"
  description   = "Function to scan S3 objects for viruses"
  image_uri     = var.ecr_image_uri
  package_type  = "Image"
  role          = var.lambda_task_role.arn
  timeout       = 300
  memory_size   = 4096
  publish       = true

  tracing_config {
    mode = "Active"
  }

  vpc_config {
    subnet_ids = var.aws_subnet_ids
    security_group_ids = [
      data.aws_security_group.lambda_egress.id
    ]
  }

  dynamic "environment" {
    for_each = length(keys(var.environment_variables)) == 0 ? [] : [true]
    content {
      variables = var.environment_variables
    }
  }
  provider = aws.region
}

data "aws_security_group" "lambda_egress" {
  name     = "lambda-egress-${data.aws_region.current.name}"
  provider = aws.region
}

resource "aws_lambda_alias" "lambda_alias" {
  name             = "latest"
  function_name    = aws_lambda_function.lambda_function.function_name
  function_version = aws_lambda_function.lambda_function.version
  provider         = aws.region
}

resource "aws_lambda_provisioned_concurrency_config" "main" {
  count                             = var.s3_antivirus_provisioned_concurrency > 0 ? 1 : 0
  function_name                     = aws_lambda_alias.lambda_alias.function_name
  provisioned_concurrent_executions = var.s3_antivirus_provisioned_concurrency
  qualifier                         = aws_lambda_alias.lambda_alias.name
  provider                          = aws.region
}
