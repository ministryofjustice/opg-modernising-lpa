module "s3_batch_manifests" {
  source                      = "./modules/s3_batch_manifests"
  s3_encryption_kms_key_alias = var.reduced_fees_uploads_s3_encryption_kms_key_alias
  providers = {
    aws.region = aws.region
  }
}
