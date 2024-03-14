#tfsec:ignore:aws-cloudwatch-log-group-customer-key
resource "aws_cloudwatch_log_group" "events" {
  name              = "${data.aws_default_tags.current.tags.environment-name}-events"
  retention_in_days = 1
  provider          = aws.region
}
