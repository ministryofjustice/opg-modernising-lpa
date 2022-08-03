resource "aws_ecs_cluster" "main" {
  name = local.name_prefix
  setting {
    name  = "containerInsights"
    value = "enabled"
  }
  provider = aws.region
}

module "app" {
  source                    = "./modules/app"
  account_name              = var.account_name
  ecs_cluster               = aws_ecs_cluster.main.id
  ecs_execution_role        = var.ecs_execution_role
  ecs_service_desired_count = 0
  network = {
    vpc_id              = data.aws_vpc.main.id
    application_subnets = data.aws_subnet.application.*.id
    public_subnets      = data.aws_subnet.public.*.id
  }
  providers = {
    aws.region = aws.region
  }
}
