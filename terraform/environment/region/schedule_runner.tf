data "aws_ecr_repository" "schedule_runner" {
  name     = "modernising-lpa/schedule-runner"
  provider = aws.management
}

module "schedule_runner" {
  source                         = "./modules/schedule_runner"
  lambda_function_image_ecr_url  = data.aws_ecr_repository.schedule_runner.repository_url
  lambda_function_image_tag      = var.app_service_container_version
  event_bus                      = module.event_bus.event_bus
  certificate_provider_start_url = var.app_env_vars.certificate_provider_start_url
  attorney_start_url             = var.app_env_vars.attorney_start_url
  search_endpoint                = var.search_endpoint
  search_index_name              = var.search_index_name
  schedule_runner_scheduler      = var.iam_roles.schedule_runner_scheduler
  schedule_runner_lambda_role    = var.iam_roles.schedule_runner_lambda
  lpa_store_base_url             = var.lpa_store_service.base_url
  app_public_url                 = aws_route53_record.app.fqdn
  allowed_api_arns               = var.lpa_store_service.api_arns
  vpc_config = {
    subnet_ids         = data.aws_subnet.application[*].id
    security_group_ids = [data.aws_security_group.lambda_egress.id]
  }

  lpas_table = {
    arn  = var.lpas_table.arn
    name = var.lpas_table.name
  }

  providers = {
    aws.region     = aws.region
    aws.management = aws.management
  }
}
