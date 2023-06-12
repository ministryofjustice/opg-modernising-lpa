moved {
  from = aws_secretsmanager_secret.rum_monitor_application_id_eu_west_1
  to   = module.eu_west_1[0].aws_secretsmanager_secret.rum_monitor_application_id
}

moved {
  from = aws_resourcegroups_group.environment-eu-west-1
  to   = module.eu_west_1[0].aws_resourcegroups_group.environment
}

moved {
  from = aws_resourcegroups_group.environment_global
  to   = module.global.aws_resourcegroups_group.environment
}
