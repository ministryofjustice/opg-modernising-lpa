data "aws_ecr_repository" "event_received" {
  name     = "modernising-lpa/event-received"
  provider = aws.management
}

module "event_received" {
  source                        = "./modules/event_received"
  lambda_function_image_ecr_url = data.aws_ecr_repository.event_received.repository_url
  lambda_function_image_tag     = var.app_service_container_version
  event_bus_name                = module.event_bus.event_bus.name
  app_public_url                = aws_route53_record.app.fqdn
  uploads_bucket                = module.uploads_s3_bucket.bucket
  uid_base_url                  = var.uid_service.base_url
  allowed_api_arns              = var.uid_service.api_arns
  search_endpoint               = var.search_endpoint
  search_collection_arn         = var.search_collection_arn

  lpas_table = {
    arn  = var.lpas_table.arn
    name = var.lpas_table.name
  }

  providers = {
    aws.region = aws.region
  }
}
