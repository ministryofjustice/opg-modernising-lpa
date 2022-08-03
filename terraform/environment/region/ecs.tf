resource "aws_ecs_cluster" "main" {
  name = local.name_prefix
  setting {
    name  = "containerInsights"
    value = "enabled"
  }
  provider = aws.region
}

# module "application_logs" {
#   source = "./modules/application_logs"
#   providers = {
#     aws.region = aws.region
#   }
# }

module "app" {
  source                    = "./modules/app"
  account_name              = var.account_name
  ecs_cluster               = aws_ecs_cluster.main.id
  ecs_execution_role        = var.ecs_execution_role
  ecs_task_role_arn         = var.ecs_task_role_arns.app
  ecs_service_desired_count = 0
  # application_log_group_arn = module.application_logs.log_group_arn
  network = {
    vpc_id              = data.aws_vpc.main.id
    application_subnets = data.aws_subnet.application.*.id
    public_subnets      = data.aws_subnet.public.*.id
  }
  providers = {
    aws.region = aws.region
  }
}
