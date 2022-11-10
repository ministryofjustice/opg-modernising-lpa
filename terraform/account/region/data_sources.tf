data "aws_kms_alias" "cloudwatch_application_logs_encryption" {
  name     = var.cloudwatch_log_group_kms_key_alias
  provider = aws.region
}

# data "aws_kms_alias" "s3_encryption" {
#   name     = var.s3_kms_key_alias
#   provider = aws.region
# }
