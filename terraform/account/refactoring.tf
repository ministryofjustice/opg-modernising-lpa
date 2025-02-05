moved {
  from = module.eu_west_1.aws_backup_vault.main
  to   = aws_backup_vault.eu_west_1
}

moved {
  from = aws_kms_key.cloudwatch
  to   = module.cloudwatch_kms.aws_kms_key.main
}

moved {
  from = aws_kms_replica_key.cloudwatch_replica
  to   = module.cloudwatch_kms.aws_kms_replica_key.main
}

moved {
  from = aws_kms_alias.cloudwatch_alias_eu_west_1
  to   = module.cloudwatch_kms.aws_kms_alias.main_eu_west_1
}

moved {
  from = aws_kms_alias.cloudwatch_alias_eu_west_2
  to   = module.cloudwatch_kms.aws_kms_alias.main_eu_west_2
}


moved {
  from = aws_kms_key.dynamodb_exports_s3_bucket
  to   = module.dynamodb_exports_s3_bucket_kms.aws_kms_key.main
}

moved {
  from = aws_kms_replica_key.dynamodb_exports_s3_bucket_replica
  to   = module.dynamodb_exports_s3_bucket_kms.aws_kms_replica_key.main
}

moved {
  from = aws_kms_alias.dynamodb_exports_s3_bucket_alias_eu_west_1
  to   = module.dynamodb_exports_s3_bucket_kms.aws_kms_alias.main_eu_west_1
}

moved {
  from = aws_kms_alias.dynamodb_exports_s3_bucket_alias_eu_west_2
  to   = module.dynamodb_exports_s3_bucket_kms.aws_kms_alias.main_eu_west_2
}


moved {
  from = aws_kms_key.dynamodb
  to   = module.dynamodb_kms.aws_kms_key.main
}

moved {
  from = aws_kms_replica_key.dynamodb_replica
  to   = module.dynamodb_kms.aws_kms_replica_key.main
}

moved {
  from = aws_kms_alias.dynamodb_alias_eu_west_1
  to   = module.dynamodb_kms.aws_kms_alias.main_eu_west_1
}

moved {
  from = aws_kms_alias.dynamodb_alias_eu_west_2
  to   = module.dynamodb_kms.aws_kms_alias.main_eu_west_2
}


moved {
  from = aws_kms_key.opensearch
  to   = module.opensearch_kms.aws_kms_key.main
}

moved {
  from = aws_kms_replica_key.opensearch_replica
  to   = module.opensearch_kms.aws_kms_replica_key.main
}

moved {
  from = aws_kms_alias.opensearch_alias_eu_west_1
  to   = module.opensearch_kms.aws_kms_alias.main_eu_west_1
}

moved {
  from = aws_kms_alias.opensearch_alias_eu_west_2
  to   = module.opensearch_kms.aws_kms_alias.main_eu_west_2
}


moved {
  from = aws_kms_key.reduced_fees_uploads_s3
  to   = module.reduced_fees_uploads_s3_kms.aws_kms_key.main
}

moved {
  from = aws_kms_replica_key.reduced_fees_uploads_s3_replica
  to   = module.reduced_fees_uploads_s3_kms.aws_kms_replica_key.main
}

moved {
  from = aws_kms_alias.reduced_fees_uploads_s3_alias_eu_west_1
  to   = module.reduced_fees_uploads_s3_kms.aws_kms_alias.main_eu_west_1
}

moved {
  from = aws_kms_alias.reduced_fees_uploads_s3_alias_eu_west_2
  to   = module.reduced_fees_uploads_s3_kms.aws_kms_alias.main_eu_west_2
}


moved {
  from = aws_kms_key.secrets_manager
  to   = module.secrets_manager_kms.aws_kms_key.main
}

moved {
  from = aws_kms_replica_key.secrets_manager_replica
  to   = module.secrets_manager_kms.aws_kms_replica_key.main
}

moved {
  from = aws_kms_alias.secrets_manager_alias_eu_west_1
  to   = module.secrets_manager_kms.aws_kms_alias.main_eu_west_1
}

moved {
  from = aws_kms_alias.secrets_manager_alias_eu_west_2
  to   = module.secrets_manager_kms.aws_kms_alias.main_eu_west_2
}


moved {
  from = aws_kms_key.sns
  to   = module.sns_kms.aws_kms_key.main
}

moved {
  from = aws_kms_replica_key.sns_replica
  to   = module.sns_kms.aws_kms_replica_key.main
}

moved {
  from = aws_kms_alias.sns_alias_eu_west_1
  to   = module.sns_kms.aws_kms_alias.main_eu_west_1
}

moved {
  from = aws_kms_alias.sns_alias_eu_west_2
  to   = module.sns_kms.aws_kms_alias.main_eu_west_2
}


moved {
  from = aws_kms_key.sqs
  to   = module.sqs_kms.aws_kms_key.main
}

moved {
  from = aws_kms_replica_key.sqs_replica
  to   = module.sqs_kms.aws_kms_replica_key.main
}

moved {
  from = aws_kms_alias.sqs_alias_eu_west_1
  to   = module.sqs_kms.aws_kms_alias.main_eu_west_1
}

moved {
  from = aws_kms_alias.sqs_alias_eu_west_2
  to   = module.sqs_kms.aws_kms_alias.main_eu_west_2
}

moved {
  from = aws_backup_vault.eu_west_1
  to   = module.eu_west_1[0].module.aws_backup_vaults.aws_backup_vault.main
}

moved {
  from = module.eu_west_1[0].module.aws_backup_vaults.aws_backup_vault.main
  to   = module.aws_backup_vaults_eu_west_1.aws_backup_vault.main
}

moved {
  from = module.eu_west_1[0].module.aws_backup_vaults.aws_backup_vault_notifications.aws_backup_failure_events
  to   = module.aws_backup_vaults_eu_west_1.aws_backup_vault_notifications.aws_backup_failure_events
}

moved {
  from = module.eu_west_1[0].module.aws_backup_vaults.aws_sns_topic.aws_backup_failure_events
  to   = module.aws_backup_vaults_eu_west_1.aws_sns_topic.aws_backup_failure_events
}

moved {
  from = module.eu_west_1[0].module.aws_backup_vaults.aws_sns_topic_policy.aws_backup_failure_events
  to   = module.aws_backup_vaults_eu_west_1.aws_sns_topic_policy.aws_backup_failure_events
}

moved {
  from = aws_cloudwatch_metric_alarm.opensearch_4xx_errors
  to   = module.eu_west_1[0].aws_cloudwatch_metric_alarm.opensearch_4xx_errors
}

moved {
  from = aws_cloudwatch_metric_alarm.opensearch_5xx_errors
  to   = module.eu_west_1[0].aws_cloudwatch_metric_alarm.opensearch_5xx_errors
}

moved {
  from = aws_opensearchserverless_access_policy.github_actions_access[0]
  to   = module.eu_west_1[0].aws_opensearchserverless_access_policy.github_actions_access[0]
}

moved {
  from = aws_opensearchserverless_access_policy.team_operator_access[0]
  to   = module.eu_west_1[0].aws_opensearchserverless_access_policy.team_operator_access[0]
}

moved {
  from = aws_opensearchserverless_collection.lpas_collection
  to   = module.eu_west_1[0].aws_opensearchserverless_collection.lpas_collection
}

moved {
  from = aws_opensearchserverless_security_policy.lpas_collection_development_network_policy[0]
  to   = module.eu_west_1[0].aws_opensearchserverless_security_policy.lpas_collection_development_network_policy[0]
}

moved {
  from = aws_opensearchserverless_security_policy.lpas_collection_network_policy
  to   = module.eu_west_1[0].aws_opensearchserverless_security_policy.lpas_collection_network_policy
}

moved {
  from = aws_sns_topic.opensearch
  to   = module.eu_west_1[0].aws_sns_topic.opensearch
}

moved {
  from = aws_opensearchserverless_security_policy.lpas_collection_encryption_policy
  to   = module.eu_west_1[0].aws_opensearchserverless_security_policy.lpas_collection_encryption_policy
}

moved {
  from = aws_sns_topic_subscription.opensearch
  to   = module.eu_west_1[0].aws_sns_topic_subscription.opensearch
}

moved {
  from = pagerduty_service_integration.opensearch
  to   = module.eu_west_1[0].pagerduty_service_integration.opensearch
}

moved {
  from = aws_opensearchserverless_access_policy.team_breakglass_access[0]
  to   = module.eu_west_1[0].aws_opensearchserverless_access_policy.team_breakglass_access[0]
}
