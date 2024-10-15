data "aws_ecr_repository" "app" {
  name     = "modernising-lpa/app"
  provider = aws.management_eu_west_1
}

data "aws_ecr_repository" "mock_onelogin" {
  name     = "mock-onelogin"
  provider = aws.management_eu_west_1
}

data "aws_ecr_repository" "mock_pay" {
  name     = "modernising-lpa/mock-pay"
  provider = aws.management_eu_west_1
}

data "aws_ecr_image" "mock_onelogin" {
  repository_name = data.aws_ecr_repository.mock_onelogin.name
  image_tag       = "latest"
  provider        = aws.management_eu_west_1
}

module "allow_list" {
  source = "git@github.com:ministryofjustice/opg-terraform-aws-moj-ip-allow-list.git?ref=v3.0.1"
}

module "eu_west_1" {
  source = "./region"
  count  = contains(local.environment.regions, "eu-west-1") ? 1 : 0
  iam_roles = {
    ecs_execution_role                      = module.global.iam_roles.ecs_execution_role
    app_ecs_task_role                       = module.global.iam_roles.app_ecs_task_role
    s3_antivirus                            = module.global.iam_roles.s3_antivirus
    cross_account_put                       = module.global.iam_roles.cross_account_put
    fault_injection_simulator               = module.global.iam_roles.fault_injection_simulator
    create_s3_batch_replication_jobs_lambda = module.global.iam_roles.create_s3_batch_replication_jobs_lambda
    event_received_lambda                   = module.global.iam_roles.event_received_lambda
    schedule_runner_lambda                  = module.global.iam_roles.schedule_runner_lambda
    schedule_runner_scheduler               = module.global.iam_roles.schedule_runner_scheduler
  }
  application_log_retention_days          = local.environment.cloudwatch_log_groups.application_log_retention_days
  ecs_capacity_provider                   = local.ecs_capacity_provider
  ecs_task_autoscaling                    = local.environment.app.autoscaling
  app_service_repository_url              = data.aws_ecr_repository.app.repository_url
  app_service_container_version           = var.container_version
  mock_onelogin_service_repository_url    = data.aws_ecr_repository.mock_onelogin.repository_url
  mock_onelogin_service_container_version = data.aws_ecr_image.mock_onelogin.id
  mock_pay_service_repository_url         = data.aws_ecr_repository.mock_pay.repository_url
  mock_pay_service_container_version      = var.container_version
  ingress_allow_list_cidr                 = module.allow_list.moj_sites
  alb_deletion_protection_enabled         = local.environment.application_load_balancer.deletion_protection_enabled
  waf_alb_association_enabled             = local.environment.application_load_balancer.waf_alb_association_enabled
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
  mock_pay_enabled                        = local.environment.mock_pay_enabled
  dependency_health_check_alarm_enabled   = local.environment.app.dependency_health_check_alarm_enabled
  service_health_check_alarm_enabled      = local.environment.app.service_health_check_alarm_enabled
  cloudwatch_application_insights_enabled = local.environment.app.cloudwatch_application_insights_enabled
  fault_injection_experiments_enabled     = local.environment.app.fault_injection_experiments_enabled
  search_endpoint                         = data.aws_opensearchserverless_collection.lpas_collection.collection_endpoint
  search_index_name                       = local.search_index_name
  search_collection_arn                   = data.aws_opensearchserverless_collection.lpas_collection.arn
  real_user_monitoring_cw_logs_enabled    = local.environment.app.real_user_monitoring_cw_logs_enabled
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
    ecs_execution_role                      = module.global.iam_roles.ecs_execution_role
    app_ecs_task_role                       = module.global.iam_roles.app_ecs_task_role
    s3_antivirus                            = module.global.iam_roles.s3_antivirus
    cross_account_put                       = module.global.iam_roles.cross_account_put
    fault_injection_simulator               = module.global.iam_roles.fault_injection_simulator
    create_s3_batch_replication_jobs_lambda = module.global.iam_roles.create_s3_batch_replication_jobs_lambda
    event_received_lambda                   = module.global.iam_roles.event_received_lambda
    schedule_runner_lambda                  = module.global.iam_roles.schedule_runner_lambda
    schedule_runner_scheduler               = module.global.iam_roles.schedule_runner_scheduler
  }
  application_log_retention_days          = local.environment.cloudwatch_log_groups.application_log_retention_days
  ecs_capacity_provider                   = local.ecs_capacity_provider
  ecs_task_autoscaling                    = local.environment.app.autoscaling
  app_service_repository_url              = data.aws_ecr_repository.app.repository_url
  app_service_container_version           = var.container_version
  mock_onelogin_service_repository_url    = data.aws_ecr_repository.mock_onelogin.repository_url
  mock_onelogin_service_container_version = local.mock_onelogin_version
  mock_pay_service_repository_url         = data.aws_ecr_repository.mock_pay.repository_url
  mock_pay_service_container_version      = var.container_version
  ingress_allow_list_cidr                 = module.allow_list.moj_sites
  alb_deletion_protection_enabled         = local.environment.application_load_balancer.deletion_protection_enabled
  waf_alb_association_enabled             = local.environment.application_load_balancer.waf_alb_association_enabled
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
  mock_pay_enabled                        = local.environment.mock_pay_enabled
  dependency_health_check_alarm_enabled   = local.environment.app.dependency_health_check_alarm_enabled
  service_health_check_alarm_enabled      = local.environment.app.service_health_check_alarm_enabled
  cloudwatch_application_insights_enabled = local.environment.app.cloudwatch_application_insights_enabled
  fault_injection_experiments_enabled     = local.environment.app.fault_injection_experiments_enabled
  search_endpoint                         = null
  search_index_name                       = local.search_index_name
  search_collection_arn                   = null
  real_user_monitoring_cw_logs_enabled    = local.environment.app.real_user_monitoring_cw_logs_enabled
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
