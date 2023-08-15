data "aws_ecr_repository" "event_received" {
  name     = "modernising-lpa/event-received"
  provider = aws.management
}

module "event_received" {
  source                        = "./modules/event_received"
  alarm_sns_topic_arn           = data.aws_sns_topic.custom_cloudwatch_alarms.arn
  aws_subnet_ids                = data.aws_subnet.application.*.id
  lambda_function_ecr_image_uri = "${data.aws_ecr_repository.event_received.repository_url}@${data.aws_ecr_image.event_received.image_digest}"
  lpas_table_name               = var.lpas_table.name

  providers = {
    aws.region = aws.region
  }
}
