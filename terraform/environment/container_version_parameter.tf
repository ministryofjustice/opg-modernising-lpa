resource "aws_ssm_parameter" "container_version" {
  name     = "/modernising-lpa/container-version/${local.environment_name}"
  type     = "String"
  value    = var.container_version
  provider = aws.management_global
}
resource "aws_ssm_parameter" "app_maintenance_switch" {
  name            = "/modernising-lpa/maintenance_mode_enabled/${data.aws_default_tags.current.tags.environment-name}"
  type            = "String"
  value           = "false"
  description     = "values of either 'true' or 'false' only"
  allowed_pattern = "^(true|false)"
  overwrite       = true
  lifecycle {
    ignore_changes = [value]
  }
  provider = aws.management_global
}
