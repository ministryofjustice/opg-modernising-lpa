data "aws_ssm_parameter" "replication_encryption_key" {
  name     = "/modernising-lpa/reduced_fees_uploads_bucket_kms_key_arn/dev/${data.aws_region.current.name}"
  provider = aws.management
}


module "uploads_s3_bucket" {
  source = "./modules/uploads_s3_bucket"

  bucket_name                           = "uploads-${data.aws_default_tags.current.tags.application}-${data.aws_default_tags.current.tags.environment-name}-${data.aws_region.current.name}"
  force_destroy                         = data.aws_default_tags.current.tags.environment-name != "production" ? true : false
  s3_replication_target_bucket_arn      = var.reduced_fees_uploads_s3_replication_target_bucket_arn
  replication_target_encryption_key_arn = data.aws_ssm_parameter.replication_encryption_key.value
  providers = {
    aws.region = aws.region
  }
}
