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

  logging_config {
    log_group  = aws_cloudwatch_log_group.lambda.name
    log_format = "JSON"
  }

  dynamic "environment" {
    for_each = length(keys(var.environment_variables)) == 0 ? [] : [true]
    content {
      variables = var.environment_variables
    }
  }
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
