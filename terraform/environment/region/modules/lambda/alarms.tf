data "aws_sns_topic" "custom_cloudwatch_alarms" {
  name     = "custom_cloudwatch_alarms"
  provider = aws.region
}

resource "aws_cloudwatch_metric_alarm" "lambda_function_failures" {
  alarm_name          = "${var.lambda_name}-${var.environment}-failures"
  comparison_operator = "GreaterThanOrEqualToThreshold"
  evaluation_periods  = "1"
  metric_name         = "Errors"
  namespace           = "AWS/Lambda"
  period              = "60"
  statistic           = "Sum"
  threshold           = "1"
  alarm_description   = "This metric monitors the number of errors that occur in the lambda function"
  alarm_actions       = [data.aws_sns_topic.custom_cloudwatch_alarms.arn]
  dimensions = {
    FunctionName = "${var.lambda_name}-${var.environment}"
  }
  provider = aws.region
}
