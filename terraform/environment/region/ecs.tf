resource "aws_ecs_cluster" "main" {
  name = data.aws_default_tags.current.tags.environment-name
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
  source                         = "./modules/app"
  ecs_cluster                    = aws_ecs_cluster.main.id
  ecs_execution_role             = var.iam_roles.ecs_execution_role
  ecs_task_role                  = var.iam_roles.app_ecs_task_role
  ecs_service_desired_count      = var.ecs_task_autoscaling.minimum
  ecs_application_log_group_name = module.application_logs.cloudwatch_log_group.name
  ecs_capacity_provider          = var.ecs_capacity_provider
  app_env_vars                   = var.app_env_vars
  app_service_repository_url     = var.app_service_repository_url
  app_service_container_version  = var.app_service_container_version
  app_allowed_api_arns = concat(
    var.uid_service.api_arns.healthcheck,
    var.lpa_store_service.api_arns.post,
    var.lpa_store_service.api_arns.put,
    var.lpa_store_service.api_arns.get,
    var.lpa_store_service.api_arns.update,
    var.lpa_store_service.api_arns.healthcheck
  )
  ingress_allow_list_cidr         = concat(var.ingress_allow_list_cidr, split(",", data.aws_ssm_parameter.additional_allowed_ingress_cidrs.value))
  alb_deletion_protection_enabled = var.alb_deletion_protection_enabled
  waf_alb_association_enabled     = var.waf_alb_association_enabled
  lpas_table                      = var.lpas_table
  sessions_table                  = var.sessions_table
  container_port                  = 8080
  public_access_enabled           = var.public_access_enabled
  network = {
    vpc_id              = data.aws_vpc.main.id
    application_subnets = data.aws_subnet.application[*].id
    public_subnets      = data.aws_subnet.public[*].id
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
  lpa_store_base_url                                   = var.lpa_store_service.base_url
  mock_onelogin_enabled                                = data.aws_default_tags.current.tags.environment-name != "production" && var.mock_onelogin.enabled
  mock_pay_enabled                                     = data.aws_default_tags.current.tags.environment-name != "production" && var.mock_pay_enabled
  fault_injection_experiments_enabled                  = var.fault_injection_experiments_enabled
  search_endpoint                                      = var.search_endpoint
  search_index_name                                    = var.search_index_name
  search_collection_arn                                = var.search_collection_arn
  ecs_aws_otel_collector_version                       = var.ecs_aws_otel_collector_version
  start_page_redirects = {
    enabled                 = var.start_page_redirects.enabled
    start_page_redirect_url = data.aws_default_tags.current.tags.environment-name != "production" ? "${data.aws_default_tags.current.tags.environment-name}.mainstreamcontent.modernising.opg.service.justice.gov.uk" : "mainstreamcontent.modernising.opg.service.justice.gov.uk"
  }

  providers = {
    aws.region     = aws.region
    aws.management = aws.management
  }
}

module "mock_onelogin" {
  count                           = data.aws_default_tags.current.tags.environment-name != "production" && var.mock_onelogin.enabled ? 1 : 0
  source                          = "./modules/mock_onelogin"
  ecs_cluster                     = aws_ecs_cluster.main.id
  ecs_execution_role              = var.iam_roles.ecs_execution_role
  ecs_task_role                   = var.iam_roles.app_ecs_task_role
  ecs_service_desired_count       = 1
  ecs_application_log_group_name  = module.application_logs.cloudwatch_log_group.name
  ecs_capacity_provider           = var.ecs_capacity_provider
  ingress_allow_list_cidr         = concat(var.ingress_allow_list_cidr, split(",", data.aws_ssm_parameter.additional_allowed_ingress_cidrs.value))
  repository_url                  = var.mock_onelogin_service_repository_url
  container_version               = var.mock_onelogin_service_container_version
  alb_deletion_protection_enabled = var.alb_deletion_protection_enabled
  waf_alb_association_enabled     = var.waf_alb_association_enabled
  container_port                  = 8080
  public_access_enabled           = var.public_access_enabled
  redirect_base_url               = var.app_env_vars.auth_redirect_base_url
  template_sub                    = var.mock_onelogin.template_sub
  network = {
    vpc_id              = data.aws_vpc.main.id
    application_subnets = data.aws_subnet.application[*].id
    public_subnets      = data.aws_subnet.public[*].id
  }
  aws_service_discovery_private_dns_namespace = {
    id   = aws_service_discovery_private_dns_namespace.internal.id
    name = aws_service_discovery_private_dns_namespace.internal.name
  }
  app_ecs_service_security_group_id = module.app.ecs_service_security_group.id
  providers = {
    aws.region = aws.region
  }
}

module "mock_pay" {
  count                           = data.aws_default_tags.current.tags.environment-name != "production" && var.mock_pay_enabled ? 1 : 0
  source                          = "./modules/mock_pay"
  ecs_cluster                     = aws_ecs_cluster.main.id
  ecs_execution_role              = var.iam_roles.ecs_execution_role
  ecs_task_role                   = var.iam_roles.app_ecs_task_role
  ecs_service_desired_count       = 1
  ecs_application_log_group_name  = module.application_logs.cloudwatch_log_group.name
  ecs_capacity_provider           = var.ecs_capacity_provider
  ingress_allow_list_cidr         = concat(var.ingress_allow_list_cidr, split(",", data.aws_ssm_parameter.additional_allowed_ingress_cidrs.value))
  repository_url                  = var.mock_pay_service_repository_url
  container_version               = var.mock_pay_service_container_version
  alb_deletion_protection_enabled = var.alb_deletion_protection_enabled
  waf_alb_association_enabled     = var.waf_alb_association_enabled
  container_port                  = 8080
  public_access_enabled           = var.public_access_enabled
  network = {
    vpc_id              = data.aws_vpc.main.id
    application_subnets = data.aws_subnet.application[*].id
    public_subnets      = data.aws_subnet.public[*].id
  }
  aws_service_discovery_private_dns_namespace = {
    id   = aws_service_discovery_private_dns_namespace.internal.id
    name = aws_service_discovery_private_dns_namespace.internal.name
  }
  app_ecs_service_security_group_id = module.app.ecs_service_security_group.id
  providers = {
    aws.region = aws.region
  }
}

module "mainstream_content" {
  source                             = "./modules/mainstream_content"
  alb_deletion_protection_enabled    = false
  mrlpa_content_container_sha_digest = var.mrlpa_content_container_sha_digest
  mrlpa_content_repository_url       = var.mrlpa_content_repository_url
  mrlpa_service_url                  = data.aws_default_tags.current.tags.environment-name != "production" ? "https://${data.aws_default_tags.current.tags.environment-name}.app.modernising.opg.service.justice.gov.uk" : "https://app.modernising.opg.service.justice.gov.uk"
  container_port                     = 3000
  ecs_application_log_group_name     = module.application_logs.cloudwatch_log_group.name
  ecs_capacity_provider              = "FARGATE_SPOT"
  ecs_cluster                        = aws_ecs_cluster.main.arn
  ecs_cpu_architecture               = "ARM64"
  ecs_execution_role                 = var.iam_roles.ecs_execution_role
  ecs_task_role                      = var.iam_roles.mainstream_content_task_role
  ecs_service_desired_count          = 1
  ingress_allow_list_cidr            = var.ingress_allow_list_cidr
  network = {
    vpc_id              = data.aws_vpc.main.id
    application_subnets = data.aws_subnet.application[*].id
    public_subnets      = data.aws_subnet.public[*].id
  }
  public_access_enabled = var.public_access_enabled


  providers = {
    aws.region = aws.region
  }
}
