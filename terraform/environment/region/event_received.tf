data "aws_ecr_repository" "event_received" {
  name     = "modernising-lpa/event-received"
  provider = aws.management
}

data "aws_security_group" "lambda_egress" {
  name     = "lambda-egress-${data.aws_region.current.name}"
  provider = aws.region
}

data "aws_secretsmanager_secret" "lpa_store_jwt_key" {
  name     = "opg-data-lpa-store/${data.aws_default_tags.current.tags.account-name}/jwt-key"
  provider = aws.management
}

module "event_received" {
  source                         = "./modules/event_received"
  lambda_function_image_ecr_url  = data.aws_ecr_repository.event_received.repository_url
  lambda_function_image_tag      = var.app_service_container_version
  event_bus_name                 = module.event_bus.event_bus.name
  event_bus_arn                  = module.event_bus.event_bus.arn
  app_public_url                 = aws_route53_record.app.fqdn
  donor_start_url                = var.app_env_vars.donor_start_url
  certificate_provider_start_url = var.app_env_vars.certificate_provider_start_url
  uploads_bucket                 = module.uploads_s3_bucket.bucket
  uid_base_url                   = var.uid_service.base_url
  lpa_store_base_url             = var.lpa_store_service.base_url
  lpa_store_secret_arn           = data.aws_secretsmanager_secret.lpa_store_jwt_key.arn
  allowed_api_arns               = concat(var.uid_service.api_arns, var.lpa_store_service.api_arns)
  search_endpoint                = var.search_endpoint
  search_index_name              = var.search_index_name
  search_collection_arn          = var.search_collection_arn
  event_received_lambda_role     = var.iam_roles.event_received_lambda
  event_bus_dead_letter_queue    = module.event_bus.event_bus_dead_letter_queue
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
