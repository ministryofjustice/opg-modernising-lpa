data "aws_ecr_repository" "app" {
  name     = "modernising-lpa/app"
  provider = aws.management_eu_west_1
}

module "allow_list" {
  source = "git@github.com:ministryofjustice/opg-terraform-aws-moj-ip-allow-list.git?ref=v2.3.0"
}

module "eu_west_1" {
  source = "./region"
  count  = contains(local.environment.regions, "eu-west-1") ? 1 : 0
  iam_roles = {
    ecs_execution_role = module.global.iam_roles.ecs_execution_role
    app_ecs_task_role  = module.global.iam_roles.app_ecs_task_role
    s3_antivirus       = module.global.iam_roles.s3_antivirus
  }
  application_log_retention_days  = local.environment.cloudwatch_log_groups.application_log_retention_days
  ecs_capacity_provider           = local.ecs_capacity_provider
  ecs_task_autoscaling            = local.environment.app.autoscaling
  app_service_repository_url      = data.aws_ecr_repository.app.repository_url
  app_service_container_version   = var.container_version
  ingress_allow_list_cidr         = module.allow_list.moj_sites
  alb_deletion_protection_enabled = local.environment.application_load_balancer.deletion_protection_enabled
  lpas_table = {
    arn  = aws_dynamodb_table.lpas_table.arn,
    name = aws_dynamodb_table.lpas_table.name
  }
  reduced_fees_table = {
    arn  = module.reduced_fees[0].dynamodb_table.arn,
    name = module.reduced_fees[0].dynamodb_table.name,
  }
  app_env_vars           = local.environment.app.env
  app_allowed_api_arns   = local.environment.app.allowed_api_arns
  public_access_enabled  = var.public_access_enabled
  pagerduty_service_name = local.environment.pagerduty_service_name
  dns_weighting          = 100
  providers = {
    aws.region            = aws.eu_west_1
    aws.global            = aws.global
    aws.management_global = aws.management_global
    aws.management        = aws.management_eu_west_1
  }
}

module "eu_west_2" {
  source = "./region"
  count  = contains(local.environment.regions, "eu-west-2") ? 1 : 0
  iam_roles = {
    ecs_execution_role = module.global.iam_roles.ecs_execution_role
    app_ecs_task_role  = module.global.iam_roles.app_ecs_task_role
    s3_antivirus       = module.global.iam_roles.s3_antivirus
  }
  application_log_retention_days  = local.environment.cloudwatch_log_groups.application_log_retention_days
  ecs_capacity_provider           = local.ecs_capacity_provider
  ecs_task_autoscaling            = local.environment.app.autoscaling
  app_service_repository_url      = data.aws_ecr_repository.app.repository_url
  app_service_container_version   = var.container_version
  ingress_allow_list_cidr         = module.allow_list.moj_sites
  alb_deletion_protection_enabled = local.environment.application_load_balancer.deletion_protection_enabled
  lpas_table = {
    arn  = local.environment.dynamodb.region_replica_enabled ? aws_dynamodb_table_replica.lpas_table[0].arn : aws_dynamodb_table.lpas_table.arn,
    name = aws_dynamodb_table.lpas_table.name
  }
  reduced_fees_table = {
    arn  = module.reduced_fees[0].dynamodb_table.arn,
    name = module.reduced_fees[0].dynamodb_table.name,
  }
  app_env_vars           = local.environment.app.env
  app_allowed_api_arns   = local.environment.app.allowed_api_arns
  public_access_enabled  = var.public_access_enabled
  pagerduty_service_name = local.environment.pagerduty_service_name
  dns_weighting          = 0
  providers = {
    aws.region            = aws.eu_west_2
    aws.global            = aws.global
    aws.management_global = aws.management_global
    aws.management        = aws.management_eu_west_2
  }
}

moved {
  from = module.eu_west_1
  to   = module.eu_west_1[0]
}
