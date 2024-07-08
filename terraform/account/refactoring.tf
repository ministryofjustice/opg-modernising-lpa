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
