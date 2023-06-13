resource "aws_route53_health_check" "health_check" {
  fqdn              = aws_route53_record.app.fqdn
  reference_name    = "${substr(data.aws_default_tags.current.tags.environment-name, 0, 20)}-health-check"
  port              = 443
  type              = "HTTPS"
  failure_threshold = 1
  request_interval  = 30
  resource_path     = "/health-check"
  measure_latency   = true
  regions           = ["us-east-1", "eu-west-1", "ap-southeast-1"]
  provider          = aws.global
  tags = {
    Name = "${data.aws_default_tags.current.tags.environment-name}-health-check-${data.aws_region.current.name}"
  }
}

resource "aws_cloudwatch_metric_alarm" "health_check" {
  alarm_description   = "${data.aws_default_tags.current.tags.environment-name} health check for ${data.aws_region.current.name}}"
  alarm_name          = "${data.aws_default_tags.current.tags.environment-name}-healthcheck-alarm-${data.aws_region.current.name}"
  actions_enabled     = false
  comparison_operator = "LessThanThreshold"
  datapoints_to_alarm = 1
  evaluation_periods  = 1
  metric_name         = "HealthCheckStatus"
  namespace           = "AWS/Route53"
  period              = 60
  statistic           = "Minimum"
  threshold           = 1
  dimensions = {
    HealthCheckId = aws_route53_health_check.health_check.id
  }

  provider = aws.global
}
