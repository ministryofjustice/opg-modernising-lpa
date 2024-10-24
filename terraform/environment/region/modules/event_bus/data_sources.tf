data "aws_sns_topic" "custom_cloudwatch_alarms" {
  name     = "custom_cloudwatch_alarms"
  provider = aws.region
}

data "aws_region" "current" {
  provider = aws.region
}

data "aws_default_tags" "current" {
  provider = aws.region
}

data "aws_caller_identity" "current" {
  provider = aws.region
}
