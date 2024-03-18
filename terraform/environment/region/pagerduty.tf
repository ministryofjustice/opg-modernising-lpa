data "pagerduty_vendor" "cloudwatch" {
  name = "Cloudwatch"
}

data "pagerduty_service" "main" {
  name = var.pagerduty_service_name
}

resource "pagerduty_service_integration" "ecs_autoscaling_alarms" {
  name    = "Modernising LPA ${data.aws_default_tags.current.tags.environment-name} ${data.aws_region.current.name} ECS AutoScaling Alarm - Maximum Reached"
  service = data.pagerduty_service.main.id
  vendor  = data.pagerduty_vendor.cloudwatch.id
}

resource "aws_sns_topic_subscription" "ecs_autoscaling_alarms" {
  topic_arn              = data.aws_sns_topic.ecs_autoscaling_alarms.arn
  protocol               = "https"
  endpoint_auto_confirms = true
  endpoint               = "https://events.pagerduty.com/integration/${pagerduty_service_integration.ecs_autoscaling_alarms.integration_key}/enqueue"
  provider               = aws.region
}

data "aws_sns_topic" "cloudwatch_application_insights" {
  name     = "cloudwatch_application_insights"
  provider = aws.region
}

resource "pagerduty_service_integration" "cloudwatch_application_insights" {
  count   = var.cloudwatch_application_insights_enabled ? 1 : 0
  name    = "Modernising LPA ${data.aws_default_tags.current.tags.environment-name} ${data.aws_region.current.name} Cloudwatch Application Insights Ops Item Alarm"
  service = data.pagerduty_service.main.id
  vendor  = data.pagerduty_vendor.cloudwatch.id
}

resource "aws_sns_topic_subscription" "cloudwatch_application_insights" {
  count                  = var.cloudwatch_application_insights_enabled ? 1 : 0
  topic_arn              = data.aws_sns_topic.cloudwatch_application_insights.arn
  protocol               = "https"
  endpoint_auto_confirms = true
  endpoint               = "https://events.pagerduty.com/integration/${pagerduty_service_integration.cloudwatch_application_insights[0].integration_key}/enqueue"
  provider               = aws.region
}

resource "aws_sns_topic" "event_alarms" {
  name                                     = "${data.aws_default_tags.current.tags.environment-name}-event-alarms"
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
  provider                                 = aws.region
}

resource "pagerduty_service_integration" "event_alarms" {
  name    = "Modernising LPA ${data.aws_default_tags.current.tags.environment-name} ${data.aws_region.current.name} Eventbridge Event Alarm"
  service = data.pagerduty_service.main.id
  vendor  = data.pagerduty_vendor.cloudwatch.id
}

resource "aws_sns_topic_subscription" "event_alarms" {
  topic_arn              = aws_sns_topic.event_alarms.arn
  protocol               = "https"
  endpoint_auto_confirms = true
  endpoint               = "https://events.pagerduty.com/integration/${pagerduty_service_integration.event_alarms.integration_key}/enqueue"
  provider               = aws.region
}
