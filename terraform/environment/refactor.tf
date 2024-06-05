moved {
  from = aws_secretsmanager_secret.rum_monitor_application_id_eu_west_1
  to   = module.eu_west_1[0].aws_secretsmanager_secret.rum_monitor_application_id
}

moved {
  from = aws_applicationinsights_application.environment_eu_west_1[0]
  to   = module.eu_west_1[0].aws_applicationinsights_application.environment
}

moved {
  from = aws_applicationinsights_application.environment_global[0]
  to   = module.global.aws_applicationinsights_application.environment_global
}

moved {
  from = aws_resourcegroups_group.environment_eu_west_1
  to   = module.eu_west_1[0].aws_resourcegroups_group.environment
}

moved {
  from = aws_resourcegroups_group.environment_global
  to   = module.global.aws_resourcegroups_group.environment_global
}

moved {
  from = aws_iam_role.execution_role
  to   = module.global.aws_iam_role.execution_role
}

moved {
  from = aws_iam_role.app_task_role
  to   = module.global.aws_iam_role.app_task_role
}

moved {
  from = aws_iam_role_policy.execution_role
  to   = module.global.aws_iam_role_policy.execution_role
}

moved {
  from = aws_cloudwatch_metric_alarm.health_check
  to   = module.eu_west_1[0].aws_cloudwatch_metric_alarm.health_check
}

moved {
  from = aws_route53_health_check.health_check
  to   = module.eu_west_1[0].aws_route53_health_check.health_check
}

moved {
  from = aws_route53_record.app
  to   = module.eu_west_1[0].aws_route53_record.app
}

moved {
  from = module.reduced_fees[0]
  to   = module.eu_west_1[0].module.events
}

moved {
  from = module.eu_west_1[0].module.events.aws_cloudwatch_event_bus.reduced_fees
  to   = module.eu_west_1[0].module.events.aws_cloudwatch_event_bus.main
}

moved {
  from = module.eu_west_1[0].module.events.aws_cloudwatch_event_archive.reduced_fees
  to   = module.eu_west_1[0].module.events.aws_cloudwatch_event_archive.main
}

moved {
  from = module.eu_west_1[0].aws_service_discovery_private_dns_namespace.mock_one_login
  to   = module.eu_west_1[0].aws_service_discovery_private_dns_namespace.internal
}
