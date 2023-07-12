module "uploads_s3_bucket" {
  source = "./modules/uploads_s3_bucket"

  bucket_name   = "${data.aws_default_tags.current.tags.application}-${data.aws_default_tags.current.tags.environment-name}-${data.aws_region.current.name}"
  force_destroy = data.aws_default_tags.current.tags.environment-name ? true : false
  providers = {
    aws.region = aws.region
  }
}
