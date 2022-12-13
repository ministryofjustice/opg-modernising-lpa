resource "aws_ssm_parameter" "container_version" {
  name     = "${local.environment_name}-container-version"
  type     = "String"
  value    = var.container_version
  provider = aws.region
}
