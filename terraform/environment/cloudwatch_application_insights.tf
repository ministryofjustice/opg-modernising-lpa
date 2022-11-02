resource "aws_applicationinsights_application" "environment_eu_west_1" {
  count               = local.environment.cloudwatch_application_insights_enabled ? 1 : 0
  resource_group_name = "${local.default_tags.environment-name}-environment-eu-west-1"
  auto_config_enabled = true
  cwe_monitor_enabled = true
  provider            = aws.eu_west_1
}

resource "aws_applicationinsights_application" "environment_eu_west_2" {
  count               = local.environment.cloudwatch_application_insights_enabled ? 1 : 0
  resource_group_name = "${local.default_tags.environment-name}-environment-eu-west-2"
  auto_config_enabled = true
  cwe_monitor_enabled = true
  provider            = aws.eu_west_2
}

resource "aws_applicationinsights_application" "environment_global" {
  count               = local.environment.cloudwatch_application_insights_enabled ? 1 : 0
  resource_group_name = "${local.default_tags.environment-name}-environment-global"
  auto_config_enabled = true
  cwe_monitor_enabled = true
  provider            = aws.global
}
