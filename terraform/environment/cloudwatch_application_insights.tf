resource "aws_applicationinsights_application" "environment_eu_west_1" {
  resource_group_name = "${local.default_tags.environment-name}-environment-eu-west-1"
  auto_config_enabled = true
  provider            = aws.eu_west_1
}

resource "aws_applicationinsights_application" "environment_eu_west_2" {
  resource_group_name = "${local.default_tags.environment-name}-environment-eu-west-2"
  auto_config_enabled = true
  provider            = aws.eu_west_2
}

resource "aws_applicationinsights_application" "environment_global" {
  resource_group_name = "${local.default_tags.environment-name}-environment-global"
  auto_config_enabled = true
  provider            = aws.global
}
