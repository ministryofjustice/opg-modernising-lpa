data "aws_ecr_repository" "event_received" {
  name     = "modernising-lpa/event-received"
  provider = aws.management
}

module "event_received" {
  source                        = "./modules/event_received"
  alarm_sns_topic_arn           = data.aws_sns_topic.custom_cloudwatch_alarms.arn
  aws_subnet_ids                = data.aws_subnet.application.*.id
  lambda_function_image_ecr_arn = data.aws_ecr_repository.event_received.arn
  lambda_function_image_ecr_url = data.aws_ecr_repository.event_received.repository_url
  lambda_function_image_tag     = var.app_service_container_version
  lpas_table_name               = var.lpas_table.name

  providers = {
    aws.region = aws.region
  }
}
