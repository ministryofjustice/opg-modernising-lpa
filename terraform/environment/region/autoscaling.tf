data "aws_iam_role" "ecs_autoscaling_service_role" {
  name     = "AWSServiceRoleForApplicationAutoScaling_ECSService"
  provider = aws.global
}

module "view_ecs_autoscaling" {
  source                           = "./modules/ecs_autoscaling"
  environment                      = local.environment_name
  aws_ecs_cluster_name             = aws_ecs_cluster.main.name
  aws_ecs_service_name             = module.app.ecs_service.name
  ecs_autoscaling_service_role_arn = data.aws_iam_role.ecs_autoscaling_service_role.arn
  ecs_task_autoscaling_minimum     = local.environment.autoscaling.view.minimum
  ecs_task_autoscaling_maximum     = local.environment.autoscaling.view.maximum
  # max_scaling_alarm_actions        = [aws_sns_topic.cloudwatch_to_pagerduty.arn]
  providers = {
    aws.region = aws.region
  }
}
