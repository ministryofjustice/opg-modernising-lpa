data "aws_sns_topic" "custom_cloudwatch_alarms" {
  name     = "custom_cloudwatch_alarms"
  provider = aws.region
}
