data "aws_ecr_repository" "app" {
  name     = "modernising-lpa/app"
  provider = aws.management_eu_west_1
}

module "allow_list" {
  source = "git@github.com:ministryofjustice/opg-terraform-aws-moj-ip-allow-list.git?ref=v1.7.1"
}

module "eu_west_1" {
  source             = "./region"
  count              = contains(local.environment.regions, "eu-west-1") ? 1 : 0
  ecs_execution_role = aws_iam_role.execution_role
  ecs_task_roles = {
    app = aws_iam_role.app_task_role
  }
  application_log_retention_days                        = local.environment.cloudwatch_log_groups.application_log_retention_days
  ecs_capacity_provider                                 = local.ecs_capacity_provider
  app_service_repository_url                            = data.aws_ecr_repository.app.repository_url
  app_service_container_version                         = var.container_version
  ingress_allow_list_cidr                               = module.allow_list.moj_sites
  alb_deletion_protection_enabled                       = local.environment.application_load_balancer.deletion_protection_enabled
  lpas_table                                            = aws_dynamodb_table.lpas_table
  app_env_vars                                          = local.environment.app.env
  public_access_enabled                                 = var.public_access_enabled
  rum_monitor_identity_pool_id_secretsmanager_secret_id = data.aws_secretsmanager_secret.rum_monitor_identity_pool_id_eu_west_1.arn
  rum_monitor_application_id_secretsmanager_secret_id   = aws_secretsmanager_secret.rum_monitor_application_id_eu_west_1.id
  providers = {
    aws.region = aws.eu_west_1
    aws.global = aws.global
  }
}

module "eu_west_2" {
  source             = "./region"
  count              = contains(local.environment.regions, "eu-west-2") ? 1 : 0
  ecs_execution_role = aws_iam_role.execution_role
  ecs_task_roles = {
    app = aws_iam_role.app_task_role
  }
  application_log_retention_days                        = local.environment.cloudwatch_log_groups.application_log_retention_days
  ecs_capacity_provider                                 = local.ecs_capacity_provider
  app_service_repository_url                            = data.aws_ecr_repository.app.repository_url
  app_service_container_version                         = var.container_version
  ingress_allow_list_cidr                               = module.allow_list.moj_sites
  alb_deletion_protection_enabled                       = local.environment.application_load_balancer.deletion_protection_enabled
  lpas_table                                            = aws_dynamodb_table.lpas_table
  app_env_vars                                          = local.environment.app.env
  public_access_enabled                                 = var.public_access_enabled
  rum_monitor_identity_pool_id_secretsmanager_secret_id = data.aws_secretsmanager_secret.rum_monitor_identity_pool_id_eu_west_1.arn # would be updated to eu_west_2 when that region exists
  rum_monitor_application_id_secretsmanager_secret_id   = aws_secretsmanager_secret.rum_monitor_application_id_eu_west_2.id
  providers = {
    aws.region = aws.eu_west_2
    aws.global = aws.global
  }
}

moved {
  from = module.eu_west_1
  to   = module.eu_west_1[0]
}
