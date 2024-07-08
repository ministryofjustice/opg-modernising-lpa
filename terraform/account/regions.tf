module "eu_west_1" {
  source                                                   = "./region"
  count                                                    = contains(local.account.regions, "eu-west-1") ? 1 : 0
  network_cidr_block                                       = "10.162.0.0/16"
  cloudwatch_log_group_kms_key_alias                       = module.cloudwatch_kms.kms_key_alias_name
  sns_kms_key_alias                                        = aws_kms_alias.sns_alias_eu_west_1.name
  secrets_manager_kms_key_alias                            = aws_kms_alias.secrets_manager_alias_eu_west_1.name
  reduced_fees_uploads_s3_encryption_kms_key_alias         = aws_kms_alias.reduced_fees_uploads_s3_alias_eu_west_1.name
  dynamodb_exports_s3_bucket_server_side_encryption_key_id = module.dynamodb_exports_s3_bucket.kms_key_alias_name
  providers = {
    aws.region     = aws.eu_west_1
    aws.management = aws.management_eu_west_1
    aws.global     = aws.global
  }
}

module "eu_west_2" {
  source                                                   = "./region"
  count                                                    = contains(local.account.regions, "eu-west-2") ? 1 : 0
  network_cidr_block                                       = "10.162.0.0/16"
  cloudwatch_log_group_kms_key_alias                       = module.cloudwatch_kms.kms_key_alias_name
  sns_kms_key_alias                                        = aws_kms_alias.sns_alias_eu_west_2.name
  secrets_manager_kms_key_alias                            = aws_kms_alias.secrets_manager_alias_eu_west_2.name
  reduced_fees_uploads_s3_encryption_kms_key_alias         = aws_kms_alias.reduced_fees_uploads_s3_alias_eu_west_2.name
  dynamodb_exports_s3_bucket_server_side_encryption_key_id = module.dynamodb_exports_s3_bucket.kms_key_alias_name
  providers = {
    aws.region     = aws.eu_west_2
    aws.management = aws.management_eu_west_2
    aws.global     = aws.global
  }
}

moved {
  from = module.eu_west_1
  to   = module.eu_west_1[0]
}
