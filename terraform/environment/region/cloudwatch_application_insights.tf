resource "aws_applicationinsights_application" "environment" {
  resource_group_name = aws_resourcegroups_group.environment.name
  auto_config_enabled = true
  cwe_monitor_enabled = true
  provider            = aws.region
}
