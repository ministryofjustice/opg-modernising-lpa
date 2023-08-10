module "s3_create_batch_replication_jobs" {
  source      = "../lambda"
  lambda_name = "create-s3-batch-replication-jobs"
  description = "Function to create and run batch replication jobs"
  environment_variables = {
    ENVIRONMENT = data.aws_default_tags.current.tags.environment-name
  }
  image_uri   = "${var.s3_replication.lambda_function_image_ecr_url}:${var.s3_replication.lambda_function_image_tag}"
  ecr_arn     = var.s3_replication.lambda_function_image_ecr_arn
  environment = data.aws_default_tags.current.tags.environment-name
  kms_key     = data.aws_kms_alias.cloudwatch_application_logs_encryption.target_key_arn
  timeout     = 900
  memory      = 1024
  providers = {
    aws.region = aws.region
  }
}

# Additional IAM permissions
resource "aws_iam_role_policy" "s3_create_batch_replication_jobs" {
  name     = "create-s3-batch-replication-jobs-${data.aws_default_tags.current.tags.environment-name}"
  role     = module.s3_create_batch_replication_jobs.lambda_role.id
  policy   = data.aws_iam_policy_document.s3_create_batch_replication_jobs.json
  provider = aws.region
}

data "aws_iam_policy_document" "s3_create_batch_replication_jobs" {
  statement {
    sid    = "GetConfiguration"
    effect = "Allow"
    resources = [
      aws_ssm_parameter.s3_batch_configuration.arn,
    ]
    actions = [
      "ssm:GetParameter",
    ]
  }
  statement {
    sid    = "CreateJob"
    effect = "Allow"
    resources = [
      "*",
    ]
    actions = [
      "s3:CreateJob",
    ]
  }
  statement {
    sid    = "Passrole"
    effect = "Allow"
    resources = [
      "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/reduced-fees-uploads-replication",
    ]
    actions = [
      "iam:GetRole",
      "iam:PassRole",
    ]
  }
  provider = aws.region
}
