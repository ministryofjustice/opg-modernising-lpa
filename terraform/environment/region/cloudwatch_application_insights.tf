resource "aws_applicationinsights_application" "environment" {
  resource_group_name = "${data.aws_default_tags.current.tags.environment-name}-environment-${data.aws_region.current.name}"
  auto_config_enabled = true
  cwe_monitor_enabled = true
  provider            = aws.region
}
