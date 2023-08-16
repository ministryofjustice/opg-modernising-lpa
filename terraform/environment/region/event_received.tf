data "aws_ecr_repository" "event_received" {
  name     = "modernising-lpa/event-received"
  provider = aws.management
}

module "event_received" {
  source                        = "./modules/event_received"
  lambda_function_image_ecr_arn = data.aws_ecr_repository.event_received.arn
  lambda_function_image_ecr_url = data.aws_ecr_repository.event_received.repository_url
  lambda_function_image_tag     = var.app_service_container_version
  lpas_table = {
    arn  = var.lpas_table.arn
    name = var.lpas_table.name
  }
  event_bus_name = var.reduced_fees.event_bus.name
  
  providers = {
    aws.region = aws.region
  }
}
