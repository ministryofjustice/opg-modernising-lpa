data "aws_sns_topic" "custom_cloudwatch_alarms" {
  name     = "custom-cloudwatch-alarms"
  provider = aws.region
}
