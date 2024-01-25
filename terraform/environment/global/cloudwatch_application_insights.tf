resource "aws_applicationinsights_application" "environment_global" {
  count               = var.cloudwatch_application_insights_enabled ? 1 : 0
  resource_group_name = aws_resourcegroups_group.environment_global.name
  auto_config_enabled = true
  cwe_monitor_enabled = true
  provider            = aws.global
}
