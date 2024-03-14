resource "aws_cloudwatch_event_rule" "ecs_failed_deployment" {
  name        = "${data.aws_default_tags.current.tags.environment-name}-capture-ecs-deployment-events"
  description = "Capture Container Task Exit Events"

  event_pattern = jsonencode(
    {
      "source" : ["aws.ecs"],
      "detail-type" : ["ECS Deployment State Change"],
      "resources" : [{ "wildcard" : "arn:aws:ecs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:service/${data.aws_default_tags.current.tags.environment-name}-${data.aws_region.current.name}/app" }],
      "detail" : {
        "eventName" : ["ERROR"],
        "eventName" : ["SERVICE_DEPLOYMENT_FAILED"]
      }
    }
  )
  provider = aws.region
}

resource "aws_cloudwatch_event_target" "ecs_failed_deployment_to_cloudwatch" {
  rule      = aws_cloudwatch_event_rule.ecs_failed_deployment.name
  target_id = "${data.aws_default_tags.current.tags.environment-name}-send-ecs-deployment-failure-events-to-cloudwatch"
  arn       = var.events_aws_cloudwatch_log_group.arn
  provider  = aws.region
}

resource "aws_cloudwatch_log_metric_filter" "ecs_failed_deployment" {
  name           = "${data.aws_default_tags.current.tags.environment-name}-ecs-failed-deployment"
  pattern        = "{ $.detail.eventName = \"SERVICE_DEPLOYMENT_FAILED\" }"
  log_group_name = var.events_aws_cloudwatch_log_group.name

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
  alarm_actions             = [data.aws_sns_topic.cloudwatch_topic.arn]
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
