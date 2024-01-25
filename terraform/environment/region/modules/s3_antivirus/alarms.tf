resource "aws_s3_bucket_metric" "virus_infections" {
  bucket = var.data_store_bucket.id
  name   = "${data.aws_default_tags.current.tags.environment-name}-virus-infections"

  filter {
    tags = {
      tostring(var.environment_variables.ANTIVIRUS_TAG_KEY) = var.environment_variables.ANTIVIRUS_TAG_VALUE_FAIL
    }
  }
  provider = aws.region
}

resource "aws_cloudwatch_metric_alarm" "virus_infections" {
  alarm_actions             = [var.alarm_sns_topic_arn]
  alarm_description         = "Possible viruses detected in ${data.aws_default_tags.current.tags.environment-name} S3 objectes"
  alarm_name                = "${data.aws_default_tags.current.tags.environment-name}-virus-infections"
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  datapoints_to_alarm       = 1
  evaluation_periods        = 1
  insufficient_data_actions = []
  metric_name               = aws_s3_bucket_metric.virus_infections.name
  namespace                 = "Monitoring"
  ok_actions                = [var.alarm_sns_topic_arn]
  period                    = 300
  statistic                 = "Sum"
  threshold                 = 1
  treat_missing_data        = "notBreaching"
  provider                  = aws.region
}
