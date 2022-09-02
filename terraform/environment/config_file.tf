resource "local_file" "environment_config" {
  content  = jsonencode(local.environment_config)
  filename = "${path.module}/environment_config.json"
}

locals {
  environment_config = {
    region                              = "eu-west-1"
    account_id                          = local.environment.account_id
    app_load_balancer_security_group_id = module.eu_west_1.app_load_balancer_security_group.id
    lpas_table                          = aws_dynamodb_table.lpas_table.name
  }
}
