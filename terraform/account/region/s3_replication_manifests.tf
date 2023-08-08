module "replication_manifests" {
  source = "./modules/s3_replication_manifests"
  providers = {
    aws.region = aws.region
  }
}
