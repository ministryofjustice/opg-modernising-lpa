resource "local_file" "environment_config" {
  content  = jsonencode(local.environment_config)
  filename = "${path.module}/environment_config.json"
}
