resource "local_file" "environment_config" {
  content  = jsonencode(local.environment_config)
  filename = "${path.module}/environment_config.json"
}

locals {
  environment_config = {
    app_load_balancer_security_group_name = module.eu_west_1.app_load_balancer_security_group.name
  }
}
