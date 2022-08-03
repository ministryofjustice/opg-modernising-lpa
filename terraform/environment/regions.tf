module "eu_west_1" {
  source       = "./region"
  account_name = local.environment.account_name
  ecs_execution_role = {
    id  = aws_iam_role.execution_role.id
    arn = aws_iam_role.execution_role.arn
  }
  ecs_task_role_arns = {
    app = aws_iam_role.app_task_role.arn
  }
  application_log_retention_days = local.environment.cloudwatch_log_groups.application_log_retention_days
  ecs_capacity_provider          = local.ecs_capacity_provider
  app_service_repository_url     = data.aws_ecr_repository.app.repository_url
  app_service_container_version  = var.container_version
  providers = {
    aws.region = aws.eu_west_1
  }
}

data "aws_ecr_repository" "app" {
  name     = "modernising-lpa/app"
  provider = aws.management_eu_west_1
}
