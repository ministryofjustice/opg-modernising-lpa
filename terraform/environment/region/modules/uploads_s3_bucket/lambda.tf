module "s3_create_batch_replication_jobs" {
  source      = "../lambda"
  lambda_name = "s3-create-batch-replication-jobs"
  description = "Function to create and run batch replication jobs"
  environment_variables = {
    ENVIRONMENT = data.aws_default_tags.current.tags.environment-name
  }
  image_uri   = "modernising-lpa/s3-create-batch-replication-jobs:${var.s3_replication.lambda_function_image_tag}"
  ecr_arn     = var.s3_replication.lambda_function_image_ecr_arn
  environment = data.aws_default_tags.current.tags.environment-name
  kms_key     = data.aws_kms_alias.cloudwatch_application_logs_encryption.target_key_arn
  timeout     = 900
  memory      = 1024
}
