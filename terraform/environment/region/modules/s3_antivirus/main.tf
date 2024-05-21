resource "aws_cloudwatch_log_group" "lambda" {
  name              = "/aws/lambda/s3-antivirus-${data.aws_default_tags.current.tags.environment-name}"
  kms_key_id        = data.aws_kms_alias.cloudwatch_application_logs_encryption.target_key_arn
  retention_in_days = 30
  provider          = aws.region
}

resource "aws_cloudwatch_query_definition" "main" {
  name            = "${data.aws_default_tags.current.tags.environment-name}/s3-antivirus"
  log_group_names = [aws_cloudwatch_log_group.lambda.name]

  query_string = <<EOF
fields @timestamp, type, record.status as status, @xrayTraceId, @message, record.metrics.initDurationMs, record.metrics.durationMs
| sort @timestamp desc
EOF
  provider     = aws.region
}
