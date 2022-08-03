resource "aws_ecs_cluster" "main" {
  name = local.name_prefix
  setting {
    name  = "containerInsights"
    value = "enabled"
  }
  provider = aws.region
}

module "app" {
  source                = "./modules/app"
  account_name          = var.account_name
  ecs_execution_role_id = var.ecs_execution_role_id
  providers = {
    aws.region = aws.region
  }
}
