resource "local_file" "environment_config" {
  content  = jsonencode(local.environment_config)
  filename = "${path.module}/environment_config.json"
}

locals {
  environment_config = {
    region                              = "eu-west-1"
    account_id                          = local.environment.account_id
    app_load_balancer_security_group_id = module.eu_west_1[0].app_load_balancer_security_group.id
  }
}

resource "aws_ssm_parameter" "container_version" {
  name  = "${local.environment_name}-container-version"
  type  = "String"
  value = var.container_version
}
