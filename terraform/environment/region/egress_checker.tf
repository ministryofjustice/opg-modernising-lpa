module "egress_checker" {
  count                         = var.egress_checker_enabled ? 0 : 1
  source                        = "./modules/egress_checker"
  lambda_function_image_ecr_url = var.egress_checker_repository_url
  lambda_function_image_tag     = var.egress_checker_container_version
  egress_checker_lambda_role    = var.iam_roles.egress_checker_lambda
  vpc_config = {
    subnet_ids         = data.aws_subnet.application[*].id
    security_group_ids = [data.aws_security_group.lambda_egress.id]
  }

  providers = {
    aws.region     = aws.region
    aws.management = aws.management
  }
}
