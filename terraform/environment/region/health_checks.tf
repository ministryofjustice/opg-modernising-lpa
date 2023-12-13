data "aws_sns_topic" "health_checks" {
  name     = "health_checks"
  provider = aws.global
}

resource "aws_route53_health_check" "service_health_check" {
  fqdn              = aws_route53_record.app.fqdn
  reference_name    = "${substr(data.aws_default_tags.current.tags.environment-name, 0, 20)}-service-hc"
  port              = 443
  type              = "HTTPS"
  failure_threshold = 1
  request_interval  = 30
  resource_path     = "/health-check/service"
  measure_latency   = true
  regions           = ["us-east-1", "eu-west-1", "ap-southeast-1"]
  tags = {
    Name = "${data.aws_default_tags.current.tags.environment-name} service health check"
  }
  provider = aws.global
}

resource "aws_cloudwatch_metric_alarm" "service_health_check" {
  alarm_description   = "${data.aws_default_tags.current.tags.environment-name} service health check for"
  alarm_name          = "${data.aws_default_tags.current.tags.environment-name}-health-check-alarm"
  alarm_actions       = [aws_sns_topic_subscription.service_health_check.arn]
  ok_actions          = [aws_sns_topic_subscription.service_health_check.arn]
  comparison_operator = "LessThanThreshold"
  datapoints_to_alarm = 1
  evaluation_periods  = 1
  metric_name         = "HealthCheckStatus"
  namespace           = "AWS/Route53"
  period              = 60
  statistic           = "Minimum"
  threshold           = 1
  dimensions = {
    HealthCheckId = aws_route53_health_check.service_health_check.id
  }

  provider = aws.global
}

resource "pagerduty_service_integration" "service_health_check" {
  name    = "Modernising LPA ${data.aws_default_tags.current.tags.environment-name} Service Health Check Alarm"
  service = data.pagerduty_service.main.id
  vendor  = data.pagerduty_vendor.cloudwatch.id
}

resource "aws_sns_topic_subscription" "service_health_check" {
  topic_arn              = data.aws_sns_topic.health_checks.arn
  protocol               = "https"
  endpoint_auto_confirms = true
  endpoint               = "https://events.pagerduty.com/integration/${pagerduty_service_integration.service_health_check.integration_key}/enqueue"
  provider               = aws.global
}

resource "aws_route53_health_check" "dependency_health_check" {
  fqdn              = aws_route53_record.app.fqdn
  reference_name    = "${substr(data.aws_default_tags.current.tags.environment-name, 0, 20)}-dependency-hc"
  port              = 443
  type              = "HTTPS"
  failure_threshold = 1
  request_interval  = 30
  resource_path     = "/health-check/dependency"
  measure_latency   = true
  regions           = ["us-east-1", "eu-west-1", "ap-southeast-1"]
  tags = {
    Name = "${data.aws_default_tags.current.tags.environment-name} dependency health check"
  }
  provider = aws.global
}

resource "aws_cloudwatch_metric_alarm" "dependency_health_check" {
  alarm_description   = "${data.aws_default_tags.current.tags.environment-name} dependency health check for}"
  alarm_name          = "${data.aws_default_tags.current.tags.environment-name}-dependency-health-check-alarm"
  alarm_actions       = [aws_sns_topic_subscription.dependency_health_check.arn]
  ok_actions          = [aws_sns_topic_subscription.dependency_health_check.arn]
  comparison_operator = "LessThanThreshold"
  datapoints_to_alarm = 1
  evaluation_periods  = 1
  metric_name         = "HealthCheckStatus"
  namespace           = "AWS/Route53"
  period              = 60
  statistic           = "Minimum"
  threshold           = 1
  dimensions = {
    HealthCheckId = aws_route53_health_check.dependency_health_check.id
  }
  provider = aws.global
}

resource "pagerduty_service_integration" "dependency_health_check" {
  name    = "Modernising LPA ${data.aws_default_tags.current.tags.environment-name} Service Health Check Alarm"
  service = data.pagerduty_service.main.id
  vendor  = data.pagerduty_vendor.cloudwatch.id
}

resource "aws_sns_topic_subscription" "dependency_health_check" {
  topic_arn              = data.aws_sns_topic.health_checks.arn
  protocol               = "https"
  endpoint_auto_confirms = true
  endpoint               = "https://events.pagerduty.com/integration/${pagerduty_service_integration.service_health_check.integration_key}/enqueue"
  provider               = aws.global
}
