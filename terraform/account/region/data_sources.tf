data "aws_kms_alias" "cloudwatch_application_logs_encryption" {
  name     = var.cloudwatch_log_group_kms_key_alias
  provider = aws.region
}

data "aws_kms_alias" "secrets_manager" {
  name     = var.secrets_manager_kms_key_alias
  provider = aws.region
}
