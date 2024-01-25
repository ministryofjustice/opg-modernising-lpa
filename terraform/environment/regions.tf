data "aws_ecr_repository" "app" {
  name     = "modernising-lpa/app"
  provider = aws.management_eu_west_1
}

data "aws_ecr_repository" "mock_onelogin" {
  name     = "modernising-lpa/mock-onelogin"
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
    cross_account_put  = module.global.iam_roles.cross_account_put
  }
  application_log_retention_days          = local.environment.cloudwatch_log_groups.application_log_retention_days
  ecs_capacity_provider                   = local.ecs_capacity_provider
  ecs_task_autoscaling                    = local.environment.app.autoscaling
  app_service_repository_url              = data.aws_ecr_repository.app.repository_url
  app_service_container_version           = var.container_version
  mock_onelogin_service_repository_url    = data.aws_ecr_repository.mock_onelogin.repository_url
  mock_onelogin_service_container_version = var.container_version
  ingress_allow_list_cidr                 = module.allow_list.moj_sites
  alb_deletion_protection_enabled         = local.environment.application_load_balancer.deletion_protection_enabled
  lpas_table = {
    arn  = aws_dynamodb_table.lpas_table.arn,
    name = aws_dynamodb_table.lpas_table.name
  }

  reduced_fees = {
    s3_object_replication_enabled             = local.environment.reduced_fees.s3_object_replication_enabled
    target_environment                        = local.environment.reduced_fees.target_environment
    destination_account_id                    = local.environment.reduced_fees.destination_account_id
    enable_s3_batch_job_replication_scheduler = local.environment.reduced_fees.enable_s3_batch_job_replication_scheduler
  }
  target_event_bus_arn                 = local.environment.event_bus.target_event_bus_arn
  receive_account_ids                  = local.environment.event_bus.receive_account_ids
  app_env_vars                         = local.environment.app.env
  public_access_enabled                = var.public_access_enabled
  pagerduty_service_name               = local.environment.pagerduty_service_name
  dns_weighting                        = 100
  s3_antivirus_provisioned_concurrency = local.environment.s3_antivirus_provisioned_concurrency
  uid_service = {
    base_url = local.environment.uid_service.base_url
    api_arns = local.environment.uid_service.api_arns
  }
  lpa_store_service = {
    base_url = local.environment.lpa_store_service.base_url
    api_arns = local.environment.lpa_store_service.api_arns
  }
  mock_onelogin_enabled                   = local.environment.mock_onelogin_enabled
  dependency_health_check_alarm_enabled   = local.environment.app.dependency_health_check_alarm_enabled
  service_health_check_alarm_enabled      = local.environment.app.service_health_check_alarm_enabled
  cloudwatch_application_insights_enabled = local.environment.app.cloudwatch_application_insights_enabled
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
    cross_account_put  = module.global.iam_roles.cross_account_put
  }
  application_log_retention_days          = local.environment.cloudwatch_log_groups.application_log_retention_days
  ecs_capacity_provider                   = local.ecs_capacity_provider
  ecs_task_autoscaling                    = local.environment.app.autoscaling
  app_service_repository_url              = data.aws_ecr_repository.app.repository_url
  app_service_container_version           = var.container_version
  mock_onelogin_service_repository_url    = data.aws_ecr_repository.mock_onelogin.repository_url
  mock_onelogin_service_container_version = var.container_version
  ingress_allow_list_cidr                 = module.allow_list.moj_sites
  alb_deletion_protection_enabled         = local.environment.application_load_balancer.deletion_protection_enabled
  lpas_table = {
    arn  = local.environment.dynamodb.region_replica_enabled ? aws_dynamodb_table_replica.lpas_table[0].arn : aws_dynamodb_table.lpas_table.arn,
    name = aws_dynamodb_table.lpas_table.name
  }

  reduced_fees = {
    s3_object_replication_enabled             = local.environment.reduced_fees.s3_object_replication_enabled
    target_environment                        = local.environment.reduced_fees.target_environment
    destination_account_id                    = local.environment.reduced_fees.destination_account_id
    enable_s3_batch_job_replication_scheduler = local.environment.reduced_fees.enable_s3_batch_job_replication_scheduler
  }
  target_event_bus_arn                 = local.environment.event_bus.target_event_bus_arn
  receive_account_ids                  = local.environment.event_bus.receive_account_ids
  app_env_vars                         = local.environment.app.env
  public_access_enabled                = var.public_access_enabled
  pagerduty_service_name               = local.environment.pagerduty_service_name
  dns_weighting                        = 0
  s3_antivirus_provisioned_concurrency = local.environment.s3_antivirus_provisioned_concurrency
  uid_service = {
    base_url = local.environment.uid_service.base_url
    api_arns = local.environment.uid_service.api_arns
  }
  lpa_store_service = {
    base_url = local.environment.lpa_store_service.base_url
    api_arns = local.environment.lpa_store_service.api_arns
  }
  mock_onelogin_enabled                   = local.environment.mock_onelogin_enabled
  dependency_health_check_alarm_enabled   = local.environment.app.dependency_health_check_alarm_enabled
  service_health_check_alarm_enabled      = local.environment.app.service_health_check_alarm_enabled
  cloudwatch_application_insights_enabled = local.environment.app.cloudwatch_application_insights_enabled
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
