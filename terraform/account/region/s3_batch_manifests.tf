module "batch_manifests" {
  source = "./modules/s3_batch_manifests"
  providers = {
    aws.region = aws.region
  }
}
