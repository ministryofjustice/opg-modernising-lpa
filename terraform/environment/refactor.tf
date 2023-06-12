moved {
  from = aws_secretsmanager_secret.rum_monitor_application_id_eu_west_1
  to   = module.eu_west_1[0].aws_secretsmanager_secret.rum_monitor_application_id
}
