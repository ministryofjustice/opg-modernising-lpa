resource "aws_cloudwatch_metric_alarm" "service_root" {
  alarm_description   = "${local.environment_name} service root health check"
  alarm_name          = "${local.environment_name}-service-root-healthcheck-alarm"
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
    HealthCheckId = aws_route53_health_check.service_root.id
  }

  provider = aws.global
}

resource "aws_route53_health_check" "service_root" {
  fqdn              = aws_route53_record.app.fqdn
  reference_name    = "${substr(local.environment_name, 0, 20)}-service-root"
  port              = 443
  type              = "HTTPS"
  failure_threshold = 1
  request_interval  = 30
  resource_path     = "/health-check"
  measure_latency   = true
  regions           = ["us-east-1", "us-west-1", "us-west-2", "eu-west-1", "ap-southeast-1", "ap-southeast-2", "ap-northeast-1", "sa-east-1"]
  provider          = aws.global
}
