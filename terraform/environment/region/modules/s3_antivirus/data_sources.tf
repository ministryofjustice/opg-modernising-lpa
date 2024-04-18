
# data "aws_region" "current" {
#   provider = aws.region
# }

data "aws_caller_identity" "current" {
  provider = aws.region
}

data "aws_default_tags" "current" {
  provider = aws.region
}

data "aws_kms_alias" "uploads_encryption_key" {
  name     = "alias/${data.aws_default_tags.current.tags.application}_reduced_fees_uploads_s3_encryption"
  provider = aws.region
}

data "aws_kms_alias" "cloudwatch_application_logs_encryption" {
  name     = "alias/${data.aws_default_tags.current.tags.application}_cloudwatch_application_logs_encryption"
  provider = aws.region
}
