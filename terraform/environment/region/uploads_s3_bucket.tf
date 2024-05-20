data "aws_ssm_parameter" "replication_encryption_key" {
  name     = "/modernising-lpa/reduced_fees_uploads_bucket_kms_key_arn/${var.reduced_fees.target_environment}/${data.aws_region.current.name}"
  provider = aws.management
}

data "aws_ssm_parameter" "replication_bucket_arn" {
  name     = "/modernising-lpa/reduced_fees_uploads_bucket_arn/${var.reduced_fees.target_environment}/${data.aws_region.current.name}"
  provider = aws.management
}

data "aws_ecr_repository" "s3_create_batch_replication_jobs" {
  name     = "modernising-lpa/create-s3-batch-replication-job"
  provider = aws.management
}

module "uploads_s3_bucket" {
  source = "./modules/uploads_s3_bucket"

  bucket_name                                      = "uploads-${data.aws_default_tags.current.tags.application}-${data.aws_default_tags.current.tags.environment-name}-${data.aws_region.current.name}"
  force_destroy                                    = data.aws_default_tags.current.tags.environment-name != "production" ? true : false
  events_received_lambda_function                  = module.event_received.lambda_function
  create_s3_batch_replication_jobs_lambda_iam_role = var.iam_roles.create_s3_batch_replication_jobs_lambda
  s3_antivirus_lambda_function                     = module.s3_antivirus.zip_lambda_function
  s3_replication = {
    enabled                                   = var.reduced_fees.s3_object_replication_enabled
    destination_bucket_arn                    = data.aws_ssm_parameter.replication_bucket_arn.value
    destination_encryption_key_arn            = data.aws_ssm_parameter.replication_encryption_key.value
    destination_account_id                    = var.reduced_fees.destination_account_id
    lambda_function_image_ecr_arn             = data.aws_ecr_repository.s3_create_batch_replication_jobs.arn
    lambda_function_image_ecr_url             = data.aws_ecr_repository.s3_create_batch_replication_jobs.repository_url
    lambda_function_image_tag                 = var.app_service_container_version
    enable_s3_batch_job_replication_scheduler = var.reduced_fees.enable_s3_batch_job_replication_scheduler
  }
  providers = {
    aws.region = aws.region
  }
}
