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
