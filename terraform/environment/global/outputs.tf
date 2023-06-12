output "resource_group_arns" {
  value = [
    aws_resourcegroups_group.environment_global.arn,
  ]
}
