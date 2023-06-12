resource "aws_applicationinsights_application" "environment_global" {
  resource_group_name = aws_resourcegroups_group.environment_global.name
  auto_config_enabled = true
  cwe_monitor_enabled = true
  provider            = aws.global
}
