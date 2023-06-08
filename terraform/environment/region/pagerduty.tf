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
