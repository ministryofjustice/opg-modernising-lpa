output "resource_group_arn" {
  value = aws_resourcegroups_group.environment_global.arn
}

output "iam_roles" {
  value = {
    ecs_execution_role                      = aws_iam_role.execution_role,
    app_ecs_task_role                       = aws_iam_role.app_task_role,
    s3_antivirus                            = aws_iam_role.s3_antivirus,
    cross_account_put                       = aws_iam_role.cross_account_put,
    fault_injection_simulator               = aws_iam_role.fault_injection_simulator,
    create_s3_batch_replication_jobs_lambda = aws_iam_role.create_s3_batch_replication_jobs_lambda
    event_received_lambda                   = aws_iam_role.event_received_lambda
    schedule_runner_lambda                  = aws_iam_role.schedule_runner_lambda
    opensearch_pipeline                     = aws_iam_role.opensearch_pipeline
    schedule_runner_scheduler               = aws_iam_role.schedule_runner_scheduler
  }
}
