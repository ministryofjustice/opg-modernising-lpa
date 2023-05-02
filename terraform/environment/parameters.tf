resource "aws_ssm_parameter" "container_version" {
  name     = "/modernising-lpa/container-version/${local.environment_name}"
  type     = "String"
  value    = var.container_version
  provider = aws.management_global
}

resource "aws_ssm_parameter" "dns_target_region" {
  provider = aws.management_global
  name     = "/modernising-lpa/dns-target-region/${local.environment_name}"
  type     = "String"
  value    = "eu-west-1"
  lifecycle {
    ignore_changes = [value]
  }
}

data "aws_ssm_parameter" "dns_target_region" {
  provider = aws.management_global
  name     = aws_ssm_parameter.dns_target_region.name
}

output "region_target" {
  value = data.aws_ssm_parameter.dns_target_region.value == "eu-west-1" ? "region one" : "region two"
}
