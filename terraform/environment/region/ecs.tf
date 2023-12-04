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

data "aws_ssm_parameter" "additional_allowed_ingress_cidrs" {
  name     = "/modernising-lpa/additional-allowed-ingress-cidrs/${data.aws_default_tags.current.tags.account-name}"
  provider = aws.management_global
}

module "app" {
  source                          = "./modules/app"
  ecs_cluster                     = aws_ecs_cluster.main.id
  ecs_execution_role              = var.iam_roles.ecs_execution_role
  ecs_task_role                   = var.iam_roles.app_ecs_task_role
  ecs_service_desired_count       = 1
  ecs_application_log_group_name  = module.application_logs.cloudwatch_log_group.name
  ecs_capacity_provider           = var.ecs_capacity_provider
  app_env_vars                    = var.app_env_vars
  app_service_repository_url      = var.app_service_repository_url
  app_service_container_version   = var.app_service_container_version
  app_allowed_api_arns            = var.uid_service.api_arns
  ingress_allow_list_cidr         = concat(var.ingress_allow_list_cidr, split(",", data.aws_ssm_parameter.additional_allowed_ingress_cidrs.value))
  alb_deletion_protection_enabled = var.alb_deletion_protection_enabled
  lpas_table                      = var.lpas_table
  container_port                  = 8080
  public_access_enabled           = var.public_access_enabled
  network = {
    vpc_id              = data.aws_vpc.main.id
    application_subnets = data.aws_subnet.application.*.id
    public_subnets      = data.aws_subnet.public.*.id
  }
  uploads_s3_bucket = {
    bucket_name = module.uploads_s3_bucket.bucket.id
    bucket_arn  = module.uploads_s3_bucket.bucket.arn
  }
  event_bus = {
    name = module.event_bus.event_bus.name
    arn  = module.event_bus.event_bus.arn
  }
  aws_rum_guest_role_arn                               = data.aws_iam_role.rum_monitor_unauthenticated.arn
  rum_monitor_application_id_secretsmanager_secret_arn = aws_secretsmanager_secret.rum_monitor_application_id.id
  uid_base_url                                         = var.uid_service.base_url
  mock_onelogin_enabled                                = data.aws_default_tags.current.tags.environment-name != "production" && var.mock_onelogin_enabled
  providers = {
    aws.region = aws.region
  }
}

module "mock_onelogin" {
  source                          = "./modules/mock_onelogin"
  ecs_cluster                     = aws_ecs_cluster.main.id
  ecs_execution_role              = var.iam_roles.ecs_execution_role
  ecs_task_role                   = var.iam_roles.app_ecs_task_role
  ecs_service_desired_count       = data.aws_default_tags.current.tags.environment-name != "production" && var.mock_onelogin_enabled ? 1 : 0
  ecs_application_log_group_name  = module.application_logs.cloudwatch_log_group.name
  ecs_capacity_provider           = var.ecs_capacity_provider
  ingress_allow_list_cidr         = concat(var.ingress_allow_list_cidr, split(",", data.aws_ssm_parameter.additional_allowed_ingress_cidrs.value))
  repository_url                  = var.mock_onelogin_service_repository_url
  container_version               = var.mock_onelogin_service_container_version
  alb_deletion_protection_enabled = var.alb_deletion_protection_enabled
  container_port                  = 8080
  public_access_enabled           = var.public_access_enabled
  redirect_base_url               = var.app_env_vars.auth_redirect_base_url
  app_security_group              = module.app.ecs_service_security_group
  network = {
    vpc_id              = data.aws_vpc.main.id
    application_subnets = data.aws_subnet.application.*.id
    public_subnets      = data.aws_subnet.public.*.id
  }
  providers = {
    aws.region = aws.region
  }
}
