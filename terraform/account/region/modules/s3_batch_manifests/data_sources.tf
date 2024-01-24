data "aws_region" "current" {
  provider = aws.region
}

data "aws_default_tags" "current" {
  provider = aws.region
}

data "aws_s3_bucket" "access_logging" {
  bucket   = "s3-access-logs-${data.aws_default_tags.current.tags.application}-${data.aws_default_tags.current.tags.account-name}-${data.aws_region.current.name}"
  provider = aws.region
}

data "aws_kms_alias" "s3_encryption_kms_key_alias" {
  name     = var.s3_encryption_kms_key_alias
  provider = aws.region
}
