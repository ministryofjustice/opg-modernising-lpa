resource "aws_applicationinsights_application" "environment" {
  resource_group_name = aws_resourcegroups_group.environment.name
  auto_config_enabled = false # temporarily disabled until the bug int he provider is resolved https://github.com/hashicorp/terraform-provider-aws/issues/27277
  cwe_monitor_enabled = true
  ops_center_enabled  = false
  depends_on = [
    aws_ecs_cluster.main,
    module.app,
  ]
  provider = aws.region
}
