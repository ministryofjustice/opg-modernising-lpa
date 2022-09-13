data "aws_ecr_repository" "app" {
  name     = "modernising-lpa/app"
  provider = aws.management_eu_west_1
}

module "allow_list" {
  source = "git@github.com:ministryofjustice/opg-terraform-aws-moj-ip-allow-list.git?ref=v1.7.1"
}

module "eu_west_1" {
  source             = "./region"
  ecs_execution_role = aws_iam_role.execution_role
  ecs_task_roles = {
    app = aws_iam_role.app_task_role
  }
  application_log_retention_days = local.environment.cloudwatch_log_groups.application_log_retention_days
  ecs_capacity_provider          = local.ecs_capacity_provider
  app_service_repository_url     = data.aws_ecr_repository.app.repository_url
  app_service_container_version  = var.container_version
  ingress_allow_list_cidr        = module.allow_list.moj_sites
  alb_enable_deletion_protection = local.environment.application_load_balancer.enable_deletion_protection
  lpas_table                     = aws_dynamodb_table.lpas_table
  app_env_vars                   = local.environment.app.env
  providers = {
    aws.region = aws.eu_west_1
  }
}
