module "egress_checker" {
  source                        = "./modules/egress_checker"
  lambda_function_image_ecr_url = "311462405659.dkr.ecr.eu-west-1.amazonaws.com/egress-checker"
  lambda_function_image_tag     = var.app_service_container_version
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
