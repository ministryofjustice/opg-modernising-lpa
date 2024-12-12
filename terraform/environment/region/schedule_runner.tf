data "aws_ecr_repository" "schedule_runner" {
  name     = "modernising-lpa/schedule-runner"
  provider = aws.management
}

module "schedule_runner" {
  source                        = "./modules/schedule_runner"
  lambda_function_image_ecr_url = data.aws_ecr_repository.schedule_runner.repository_url
  lambda_function_image_tag     = var.app_service_container_version
  event_bus                     = module.event_bus.event_bus
  search_endpoint               = var.search_endpoint
  search_index_name             = var.search_index_name
  schedule_runner_scheduler     = var.iam_roles.schedule_runner_scheduler
  schedule_runner_lambda_role   = var.iam_roles.schedule_runner_lambda
  lpa_store_base_url            = var.lpa_store_service.base_url
  lpa_store_secret_arn          = data.aws_secretsmanager_secret.lpa_store_jwt_key.arn
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
