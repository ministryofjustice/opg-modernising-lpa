output "resource_group_arn" {
  value = aws_resourcegroups_group.environment_global.arn
}

output "iam_roles" {
  value = {
    ecs_execution_role = aws_iam_role.execution_role,
    app_ecs_task_role  = aws_iam_role.app_task_role,
    s3_antivirus       = aws_iam_role.s3_antivirus,
  }
}
