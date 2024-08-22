resource "aws_sns_topic" "service_health_checks_global" {
  name                                     = "${data.aws_default_tags.current.tags.environment-name}-service-health-checks"
  kms_master_key_id                        = data.aws_kms_alias.sns_kms_key_alias_global.target_key_id
  application_failure_feedback_role_arn    = data.aws_iam_role.sns_failure_feedback.arn
  application_success_feedback_role_arn    = data.aws_iam_role.sns_success_feedback.arn
  application_success_feedback_sample_rate = 100
  firehose_failure_feedback_role_arn       = data.aws_iam_role.sns_failure_feedback.arn
  firehose_success_feedback_role_arn       = data.aws_iam_role.sns_success_feedback.arn
  firehose_success_feedback_sample_rate    = 100
  http_failure_feedback_role_arn           = data.aws_iam_role.sns_failure_feedback.arn
  http_success_feedback_role_arn           = data.aws_iam_role.sns_success_feedback.arn
  http_success_feedback_sample_rate        = 100
  lambda_failure_feedback_role_arn         = data.aws_iam_role.sns_failure_feedback.arn
  lambda_success_feedback_role_arn         = data.aws_iam_role.sns_success_feedback.arn
  lambda_success_feedback_sample_rate      = 100
  sqs_failure_feedback_role_arn            = data.aws_iam_role.sns_failure_feedback.arn
  sqs_success_feedback_role_arn            = data.aws_iam_role.sns_success_feedback.arn
  sqs_success_feedback_sample_rate         = 100
  provider                                 = aws.global
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
  regions           = ["us-east-1", "eu-west-1", "us-west-2"]
  tags = {
    Name = "${data.aws_default_tags.current.tags.environment-name} service health check"
  }
  provider = aws.global
}

resource "aws_cloudwatch_metric_alarm" "service_health_check" {
  alarm_description   = "${data.aws_default_tags.current.tags.environment-name} service health check for"
  alarm_name          = "${data.aws_default_tags.current.tags.environment-name}-service-health-check-alarm"
  alarm_actions       = [aws_sns_topic.service_health_checks_global.arn]
  ok_actions          = [aws_sns_topic.service_health_checks_global.arn]
  actions_enabled     = var.service_health_check_alarm_enabled
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
  topic_arn              = aws_sns_topic.service_health_checks_global.arn
  protocol               = "https"
  endpoint_auto_confirms = true
  endpoint               = "https://events.pagerduty.com/integration/${pagerduty_service_integration.service_health_check.integration_key}/enqueue"
  provider               = aws.global
}

resource "aws_sns_topic" "dependency_health_checks_global" {
  name                                     = "${data.aws_default_tags.current.tags.environment-name}-dependency-health-checks"
  kms_master_key_id                        = data.aws_kms_alias.sns_kms_key_alias_global.target_key_id
  application_failure_feedback_role_arn    = data.aws_iam_role.sns_failure_feedback.arn
  application_success_feedback_role_arn    = data.aws_iam_role.sns_success_feedback.arn
  application_success_feedback_sample_rate = 100
  firehose_failure_feedback_role_arn       = data.aws_iam_role.sns_failure_feedback.arn
  firehose_success_feedback_role_arn       = data.aws_iam_role.sns_success_feedback.arn
  firehose_success_feedback_sample_rate    = 100
  http_failure_feedback_role_arn           = data.aws_iam_role.sns_failure_feedback.arn
  http_success_feedback_role_arn           = data.aws_iam_role.sns_success_feedback.arn
  http_success_feedback_sample_rate        = 100
  lambda_failure_feedback_role_arn         = data.aws_iam_role.sns_failure_feedback.arn
  lambda_success_feedback_role_arn         = data.aws_iam_role.sns_success_feedback.arn
  lambda_success_feedback_sample_rate      = 100
  sqs_failure_feedback_role_arn            = data.aws_iam_role.sns_failure_feedback.arn
  sqs_success_feedback_role_arn            = data.aws_iam_role.sns_success_feedback.arn
  sqs_success_feedback_sample_rate         = 100
  provider                                 = aws.global
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
  regions           = ["us-east-1", "eu-west-1", "us-west-2"]
  tags = {
    Name = "${data.aws_default_tags.current.tags.environment-name} dependency health check"
  }
  provider = aws.global
}

resource "aws_cloudwatch_metric_alarm" "dependency_health_check" {
  alarm_description   = "${data.aws_default_tags.current.tags.environment-name} dependency health check for}"
  alarm_name          = "${data.aws_default_tags.current.tags.environment-name}-dependency-health-check-alarm"
  alarm_actions       = [aws_sns_topic.dependency_health_checks_global.arn]
  ok_actions          = [aws_sns_topic.dependency_health_checks_global.arn]
  actions_enabled     = var.dependency_health_check_alarm_enabled
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
  name    = "Modernising LPA ${data.aws_default_tags.current.tags.environment-name} Dependency Health Check Alarm"
  service = data.pagerduty_service.main.id
  vendor  = data.pagerduty_vendor.cloudwatch.id
}

resource "aws_sns_topic_subscription" "dependency_health_check" {
  topic_arn              = aws_sns_topic.dependency_health_checks_global.arn
  protocol               = "https"
  endpoint_auto_confirms = true
  endpoint               = "https://events.pagerduty.com/integration/${pagerduty_service_integration.service_health_check.integration_key}/enqueue"
  provider               = aws.global
}
