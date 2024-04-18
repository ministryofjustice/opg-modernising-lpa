data "aws_ecr_repository" "s3_antivirus" {
  name     = "s3-antivirus"
  provider = aws.management
}

data "aws_ecr_image" "s3_antivirus" {
  repository_name = data.aws_ecr_repository.s3_antivirus.name
  image_tag       = "latest"
  provider        = aws.management
}

data "aws_s3_bucket" "antivirus_definitions" {
  bucket   = "virus-definitions-${data.aws_default_tags.current.tags.application}-${data.aws_default_tags.current.tags.account-name}-${data.aws_region.current.name}"
  provider = aws.region
}

module "s3_antivirus" {
  source                               = "./modules/s3_antivirus"
  alarm_sns_topic_arn                  = data.aws_sns_topic.custom_cloudwatch_alarms.arn
  data_store_bucket                    = module.uploads_s3_bucket.bucket
  definition_bucket                    = data.aws_s3_bucket.antivirus_definitions
  ecr_image_uri                        = "${data.aws_ecr_repository.s3_antivirus.repository_url}@${data.aws_ecr_image.s3_antivirus.image_digest}"
  lambda_task_role                     = var.iam_roles.s3_antivirus
  s3_antivirus_provisioned_concurrency = var.s3_antivirus_provisioned_concurrency

  environment_variables = {
    ANTIVIRUS_DEFINITIONS_BUCKET = data.aws_s3_bucket.antivirus_definitions.id
    ANTIVIRUS_TAG_KEY            = "virus-scan-status"
    ANTIVIRUS_TAG_VALUE_PASS     = "ok"
    ANTIVIRUS_TAG_VALUE_FAIL     = "infected"
  }
  providers = {
    aws.region = aws.region
  }
}

data "aws_sns_topic" "custom_cloudwatch_alarms" {
  name     = "custom_cloudwatch_alarms"
  provider = aws.region
}
