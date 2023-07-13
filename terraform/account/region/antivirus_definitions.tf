data "aws_ecr_repository" "s3_antivirus_update" {
  name     = "s3-antivirus-update"
  provider = aws.management
}

module "antivirus_definitions" {
  source        = "./modules/antivirus_definitions"
  ecr_image_uri = "${data.aws_ecr_repository.s3_antivirus_update.repository_url}:latest"
  providers = {
    aws.region = aws.region
  }
}
