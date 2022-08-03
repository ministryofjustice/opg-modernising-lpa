module "eu_west_1" {
  source                = "./region"
  account_name          = local.environment.account_name
  ecs_execution_role_id = aws_iam_role.execution_role.id
  providers = {
    aws.region = aws.eu_west_1
  }
}

data "aws_ecr_repository" "app" {
  name     = "modernising-lpa/app"
  provider = aws.management_eu_west_1
}
