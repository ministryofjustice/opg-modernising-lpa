resource "aws_lambda_layer_version" "lambda_layer" {
  filename                 = "${path.module}/lambda_layer.zip"
  layer_name               = "clamav-${data.aws_default_tags.current.tags.environment-name}"
  description              = "ClamAV Antivirus Layer"
  source_code_hash         = filebase64sha256("${path.module}/lambda_layer.zip")
  compatible_architectures = ["x86_64"]
  compatible_runtimes      = ["provided.al2023"]
  provider                 = aws.region
}

resource "terraform_data" "replacement" {
  triggers_replace = [filebase64sha256("${path.module}/myFunction.zip")]
}

resource "aws_lambda_function" "lambda_function" {
  function_name    = "zip-s3-antivirus-${data.aws_default_tags.current.tags.environment-name}"
  description      = "Function to scan S3 objects for viruses"
  filename         = "${path.module}/myFunction.zip"
  handler          = "bootstrap"
  source_code_hash = filebase64sha256("${path.module}/myFunction.zip")
  architectures    = ["x86_64"]
  runtime          = "provided.al2023"
  timeout          = 300
  memory_size      = 4096
  publish          = filebase64sha256("${path.module}/myFunction.zip") != terraform_data.replacement.triggers_replace[0] ? true : false

  layers = [
    aws_lambda_layer_version.lambda_layer.arn
  ]

  role = var.lambda_task_role.arn

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
