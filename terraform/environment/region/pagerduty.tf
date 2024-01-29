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
  name    = "Modernising LPA ${data.aws_default_tags.current.tags.environment-name} ${data.aws_region.current.name} Cloudwatch Application Insights Ops Item Alarm"
  service = data.pagerduty_service.main.id
  vendor  = data.pagerduty_vendor.cloudwatch.id
}

resource "aws_sns_topic_subscription" "cloudwatch_application_insights" {
  topic_arn              = data.aws_sns_topic.cloudwatch_application_insights.arn
  protocol               = "https"
  endpoint_auto_confirms = true
  endpoint               = "https://events.pagerduty.com/integration/${pagerduty_service_integration.ecs_autoscaling_alarms.integration_key}/enqueue"
  provider               = aws.region
}
