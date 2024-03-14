resource "aws_cloudwatch_log_group" "events" {
  name              = "${data.aws_default_tags.current.tags.environment-name}-events"
  retention_in_days = 1
  # kms_key_id        = data.aws_kms_alias.cloudwatch_application_logs_encryption.target_key_arn
  provider = aws.region
}
