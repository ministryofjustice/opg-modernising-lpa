data "aws_iam_role" "ecs_autoscaling_service_role" {
  name     = "AWSServiceRoleForApplicationAutoScaling_ECSService"
  provider = aws.global
}
