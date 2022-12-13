resource "aws_ssm_parameter" "container_version" {
  name     = "modernising-lpa/container-version/${local.environment_name}"
  type     = "String"
  value    = var.container_version
  provider = aws.management_global
}
