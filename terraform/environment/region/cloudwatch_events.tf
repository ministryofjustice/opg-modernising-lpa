#tfsec:ignore:aws-cloudwatch-log-group-customer-key
resource "aws_cloudwatch_log_group" "events" {
  name              = "/aws/events/${data.aws_default_tags.current.tags.environment-name}"
  retention_in_days = 1
  provider          = aws.region
}

resource "aws_cloudwatch_log_data_protection_policy" "events" {
  log_group_name = aws_cloudwatch_log_group.events.name
  policy_document = jsonencode(merge(
    jsondecode(file("${path.root}/cloudwatch_log_data_protection_policy/cloudwatch_log_data_protection_policy.json")),
    {
      Name = "data-protection-${data.aws_default_tags.current.tags.environment-name}-events"
    }
  ))
  provider = aws.region
}

resource "aws_cloudwatch_query_definition" "events" {
  name            = "${data.aws_default_tags.current.tags.environment-name}/events"
  log_group_names = [aws_cloudwatch_log_group.events.name]

  query_string = <<EOF
fields @timestamp, detail.eventName, detail.reason
| sort @timestamp desc
| limit 1000
EOF
  provider     = aws.region
}

resource "aws_cloudwatch_event_rule" "ecs_failed_deployment" {
  name        = "${data.aws_default_tags.current.tags.environment-name}-capture-ecs-deployment-events"
  description = "Capture ECS deployment failure events for ${data.aws_default_tags.current.tags.environment-name}"

  event_pattern = jsonencode(
    {
      "source" : ["aws.ecs"],
      "detail-type" : ["ECS Deployment State Change"],
      "resources" : [{ "wildcard" : "arn:aws:ecs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:service/${data.aws_default_tags.current.tags.environment-name}-${data.aws_region.current.name}/*" }],
      "detail" : {
        "eventType" : ["ERROR"],
        "eventName" : ["SERVICE_DEPLOYMENT_FAILED"]
      }
    }
  )
  provider = aws.region
}

resource "aws_cloudwatch_event_target" "ecs_failed_deployment_to_cloudwatch" {
  rule      = aws_cloudwatch_event_rule.ecs_failed_deployment.name
  target_id = "${data.aws_default_tags.current.tags.environment-name}-send-ecs-deployment-failure-events-to-log-group"
  arn       = aws_cloudwatch_log_group.events.arn
  provider  = aws.region
}

resource "aws_cloudwatch_log_metric_filter" "ecs_failed_deployment" {
  name           = "${data.aws_default_tags.current.tags.environment-name}-ecs-failed-deployment"
  pattern        = "{ $.detail.eventName = \"SERVICE_DEPLOYMENT_FAILED\" }"
  log_group_name = aws_cloudwatch_log_group.events.name

  metric_transformation {
    name          = "${data.aws_default_tags.current.tags.environment-name}-ecs-failed-deployment"
    namespace     = "Monitoring"
    value         = "1"
    default_value = "0"
  }
  provider = aws.region
}

resource "aws_cloudwatch_metric_alarm" "ecs_failed_deployment" {
  actions_enabled           = true
  alarm_actions             = [aws_sns_topic.event_alarms.arn]
  alarm_description         = "ECS Deployment Failure for ${data.aws_default_tags.current.tags.environment-name}"
  alarm_name                = "${data.aws_default_tags.current.tags.environment-name}-ecs-failed-deployments"
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  datapoints_to_alarm       = 1
  evaluation_periods        = 1
  insufficient_data_actions = []
  metric_name               = aws_cloudwatch_log_metric_filter.ecs_failed_deployment.name
  namespace                 = "Monitoring"
  period                    = 60
  statistic                 = "Maximum"
  threshold                 = 1
  treat_missing_data        = "notBreaching"
  provider                  = aws.region
}
