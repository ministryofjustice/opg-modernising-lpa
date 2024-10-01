resource "aws_cloudwatch_log_group" "lambda" {
  name       = "/aws/lambda/${var.environment}-${var.lambda_name}"
  kms_key_id = var.kms_key
  provider   = aws.region
}

resource "aws_cloudwatch_log_data_protection_policy" "lambda" {
  log_group_name = aws_cloudwatch_log_group.lambda.name
  policy_document = jsonencode(merge(
    jsondecode(file("${path.root}/cloudwatch_log_data_protection_policy/cloudwatch_log_data_protection_policy.json")),
    {
      Name = "data-protection-${var.environment}-${var.lambda_name}"
    }
  ))
  provider = aws.region
}

resource "aws_lambda_function" "lambda_function" {
  function_name = "${var.lambda_name}-${var.environment}"
  description   = var.description
  image_uri     = var.image_uri
  architectures = var.architectures
  package_type  = var.package_type
  role          = var.aws_iam_role.arn
  timeout       = var.timeout
  memory_size   = var.memory
  depends_on    = [aws_cloudwatch_log_group.lambda]

  tracing_config {
    mode = "Active"
  }

  logging_config {
    log_group  = aws_cloudwatch_log_group.lambda.name
    log_format = "JSON"
  }

  dynamic "vpc_config" {
    for_each = length(var.vpc_config) == 0 ? [] : [true]
    content {
      subnet_ids         = var.vpc_config.subnet_ids
      security_group_ids = var.vpc_config.security_group_ids
    }
  }

  dynamic "environment" {
    for_each = length(keys(var.environment_variables)) == 0 ? [] : [true]
    content {
      variables = var.environment_variables
    }
  }
  provider = aws.region
}

resource "aws_cloudwatch_query_definition" "main" {
  name            = "${var.environment}/${var.lambda_name}"
  log_group_names = [aws_cloudwatch_log_group.lambda.name]

  query_string = <<EOF
fields @timestamp, type, record.status as status, @xrayTraceId, @message, record.metrics.initDurationMs, record.metrics.durationMs
| sort @timestamp desc
EOF
  provider     = aws.region
}
