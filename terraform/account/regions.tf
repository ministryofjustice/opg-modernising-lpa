module "eu_west_1" {
  source                                                   = "./region"
  count                                                    = contains(local.account.regions, "eu-west-1") ? 1 : 0
  network_cidr_block                                       = "10.162.0.0/16"
  network_firewall_rules_file                              = "../account/network_firewall_rules.rules"
  cloudwatch_log_group_kms_key_alias                       = module.cloudwatch_kms.kms_key_alias_name
  sns_kms_key_alias                                        = module.sns_kms.kms_key_alias_name
  secrets_manager_kms_key_alias                            = module.secrets_manager_kms.kms_key_alias_name
  reduced_fees_uploads_s3_encryption_kms_key_alias         = module.reduced_fees_uploads_s3_kms.kms_key_alias_name
  dynamodb_exports_s3_bucket_server_side_encryption_key_id = module.dynamodb_exports_s3_bucket_kms.eu_west_1_target_key_id
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
  network_firewall_rules_file                              = "../account/network_firewall_rules.rules"
  cloudwatch_log_group_kms_key_alias                       = module.cloudwatch_kms.kms_key_alias_name
  sns_kms_key_alias                                        = module.sns_kms.kms_key_alias_name
  secrets_manager_kms_key_alias                            = module.secrets_manager_kms.kms_key_alias_name
  reduced_fees_uploads_s3_encryption_kms_key_alias         = module.reduced_fees_uploads_s3_kms.kms_key_alias_name
  dynamodb_exports_s3_bucket_server_side_encryption_key_id = module.dynamodb_exports_s3_bucket_kms.eu_west_2_target_key_id
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
