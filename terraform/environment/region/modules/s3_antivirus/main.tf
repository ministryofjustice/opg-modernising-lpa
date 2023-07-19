data "aws_kms_alias" "cloudwatch_application_logs_encryption" {
  name     = "alias/${data.aws_default_tags.current.tags.application}_cloudwatch_application_logs_encryption"
  provider = aws.region
}

resource "aws_cloudwatch_log_group" "lambda" {
  name              = "/aws/lambda/s3-antivirus-${data.aws_default_tags.current.tags.environment-name}"
  kms_key_id        = data.aws_kms_alias.cloudwatch_application_logs_encryption.target_key_id
  retention_in_days = 30
  provider          = aws.region
}

resource "aws_lambda_function" "lambda_function" {
  function_name = "s3-antivirus-${data.aws_default_tags.current.tags.environment-name}"
  image_uri     = var.ecr_image_uri
  package_type  = "Image"
  role          = var.lambda_task_role.arn
  timeout       = 300
  memory_size   = 4096

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

resource "aws_lambda_permission" "allow_bucket_to_run" {
  statement_id   = "AllowExecutionFromS3Bucket"
  action         = "lambda:InvokeFunction"
  function_name  = aws_lambda_function.lambda_function.function_name
  principal      = "s3.amazonaws.com"
  source_account = data.aws_caller_identity.current.account_id
  source_arn     = var.data_store_bucket.arn
  provider       = aws.region
}

data "aws_security_group" "lambda_egress" {
  name     = "lambda-egress-${data.aws_region.current.name}"
  provider = aws.region
}
