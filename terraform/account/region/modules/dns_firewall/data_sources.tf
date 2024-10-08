data "aws_kms_alias" "cloudwatch_application_logs_encryption" {
  name     = var.cloudwatch_log_group_kms_key_alias
  provider = aws.region
}

data "aws_region" "current" {
  provider = aws.region
}

data "aws_default_tags" "current" {
  provider = aws.region
}
