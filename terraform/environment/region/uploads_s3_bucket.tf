data "aws_ssm_parameter" "replication_encryption_key" {
  name     = "/modernising-lpa/reduced_fees_uploads_bucket_kms_key_arn/${var.reduced_fees.target_environment}/${data.aws_region.current.name}"
  provider = aws.management
}

data "aws_ssm_parameter" "replication_bucket_arn" {
  name     = "/modernising-lpa/reduced_fees_uploads_bucket_arn/${var.reduced_fees.target_environment}/${data.aws_region.current.name}"
  provider = aws.management
}

module "uploads_s3_bucket" {
  source = "./modules/uploads_s3_bucket"

  bucket_name   = "uploads-${data.aws_default_tags.current.tags.application}-${data.aws_default_tags.current.tags.environment-name}-${data.aws_region.current.name}"
  force_destroy = data.aws_default_tags.current.tags.environment-name != "production" ? true : false
  s3_replication = {
    enabled                        = var.reduced_fees.s3_object_replication_enabled
    destination_bucket_arn         = data.aws_ssm_parameter.replication_bucket_arn.value
    destination_encryption_key_arn = data.aws_ssm_parameter.replication_encryption_key.value
    destination_account_id         = var.reduced_fees.destination_account_id
  }
  providers = {
    aws.region = aws.region
  }
}
