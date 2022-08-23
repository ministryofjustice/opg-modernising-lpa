resource "local_file" "cluster_config" {
  content  = jsonencode(local.cluster_config)
  filename = "${path.module}/cluster_config.json"
}

locals {
  cluster_config = {
    app_load_balancer_security_group_name  = module.eu_west_1.app_load_balancer_security_group.name
  }
}
