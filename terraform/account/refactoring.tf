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
  to   = module.dynamodb_exports_s3_bucket.aws_kms_key.main
}

moved {
  from = aws_kms_replica_key.dynamodb_exports_s3_bucket_replica
  to   = module.dynamodb_exports_s3_bucket.aws_kms_replica_key.main
}

moved {
  from = aws_kms_alias.dynamodb_exports_s3_bucket_alias_eu_west_1
  to   = module.dynamodb_exports_s3_bucket.aws_kms_alias.main_eu_west_1
}

moved {
  from = aws_kms_alias.dynamodb_exports_s3_bucket_alias_eu_west_2
  to   = module.dynamodb_exports_s3_bucket.aws_kms_alias.main_eu_west_2
}
