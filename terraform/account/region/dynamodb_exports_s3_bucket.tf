module "dynamodb_exports_s3_bucket" {
  source                                  = "./modules/dynamodb_exports_s3_bucket"
  s3_bucket_server_side_encryption_key_id = var.dynamodb_exports_s3_bucket_server_side_encryption_key_id
  providers = {
    aws.region = aws.region
  }
}
