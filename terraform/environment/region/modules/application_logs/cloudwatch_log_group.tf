data "aws_kms_alias" "cloudwatch_application_logs_encryption" {
  name     = "alias/${data.aws_default_tags.current.tags.application}_cloudwatch_application_logs_encryption"
  provider = aws.region
}

resource "aws_cloudwatch_log_group" "application_logs" {
  name              = "${data.aws_default_tags.current.tags.environment-name}_application_logs"
  retention_in_days = var.application_log_retention_days
  kms_key_id        = data.aws_kms_alias.cloudwatch_application_logs_encryption.target_key_arn
  provider          = aws.region
}

resource "aws_cloudwatch_log_data_protection_policy" "application_logs" {
  log_group_name = aws_cloudwatch_log_group.application_logs.name
  policy_document = jsonencode(merge(
    jsondecode(file("${path.root}/cloudwatch_log_data_protection_policy/cloudwatch_log_data_protection_policy.json")),
    {
      Name = "data-protection-${data.aws_default_tags.current.tags.environment-name}-application-logs"
    }
  ))
  provider = aws.region
}

resource "aws_cloudwatch_query_definition" "app_container_messages" {
  name            = "${data.aws_default_tags.current.tags.environment-name}/app container messages"
  log_group_names = [aws_cloudwatch_log_group.application_logs.name]

  query_string = <<EOF
fields @timestamp, level, msg, err, concat(req.method, " ", req.uri) as request
| filter @message not like "ELB-HealthChecker"
| filter @logStream not like /(?i)(mock_onelogin|aws-otel-collector)/
| sort @timestamp desc
EOF
  provider     = aws.region
}
