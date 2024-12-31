module "egress_checker" {
  count                         = 0
  source                        = "./modules/egress_checker"
  lambda_function_image_ecr_url = var.egress_checker_repository_url
  lambda_function_image_tag     = var.egress_checker_container_version
  event_received_lambda_role    = var.iam_roles.event_received_lambda
  vpc_config = {
    subnet_ids         = data.aws_subnet.application[*].id
    security_group_ids = [data.aws_security_group.lambda_egress.id]
  }

  providers = {
    aws.region     = aws.region
    aws.management = aws.management
  }
}
