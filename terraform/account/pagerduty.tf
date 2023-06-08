data "pagerduty_vendor" "cloudwatch" {
  name = "Cloudwatch"
}

data "pagerduty_service" "main" {
  name = local.account.pagerduty_service_name
}

resource "pagerduty_service_integration" "main" {
  name    = "Modernising LPA ${data.pagerduty_vendor.cloudwatch.name} ${local.account.account_name} Account alerts"
  service = data.pagerduty_service.main.id
  vendor  = data.pagerduty_vendor.cloudwatch.id
}

# resource "aws_sns_topic_subscription" "main" {
#   topic_arn              = module.eu_west_1.ecs_autoscaling_alarm_sns_topic.arn
#   protocol               = "https"
#   endpoint_auto_confirms = true
#   endpoint               = "https://events.pagerduty.com/integration/${pagerduty_service_integration.main.integration_key}/enqueue"
#   provider               = aws.eu_west_1
# }

output "ecs_autoscaling_alarm_sns_topic" {
  value = module.eu_west_1.ecs_autoscaling_alarm_sns_topic
}
