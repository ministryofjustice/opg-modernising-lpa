resource "aws_ecs_cluster" "main" {
  name = local.name_prefix
  setting {
    name  = "containerInsights"
    value = "enabled"
  }
  provider = aws.region
}

module "application_logs" {
  source                         = "./modules/application_logs"
  application_log_retention_days = var.application_log_retention_days
  providers = {
    aws.region = aws.region
  }
}

module "app" {
  source                         = "./modules/app"
  ecs_cluster                    = aws_ecs_cluster.main.id
  ecs_execution_role             = var.ecs_execution_role
  ecs_task_role_arn              = var.ecs_task_role_arns.app
  ecs_service_desired_count      = 0
  ecs_application_log_group_name = module.application_logs.cloudwatch_log_group.name
  ecs_capacity_provider          = var.ecs_capacity_provider
  app_service_repository_url     = var.app_service_repository_url
  app_service_container_version  = var.app_service_container_version
  network = {
    vpc_id              = data.aws_vpc.main.id
    application_subnets = data.aws_subnet.application.*.id
    public_subnets      = data.aws_subnet.public.*.id
  }
  providers = {
    aws.region = aws.region
  }
}
