module "s3_create_batch_replication_jobs" {
  source      = "../lambda"
  lambda_name = "create-s3-batch-replication-jobs"
  description = "Function to create and run batch replication jobs in ${data.aws_region.current.name}"
  environment_variables = {
    ENVIRONMENT = data.aws_default_tags.current.tags.environment-name
  }
  image_uri    = "${var.s3_replication.lambda_function_image_ecr_url}:${var.s3_replication.lambda_function_image_tag}"
  aws_iam_role = var.create_s3_batch_replication_jobs_lambda_iam_role
  environment  = data.aws_default_tags.current.tags.environment-name
  kms_key      = data.aws_kms_alias.cloudwatch_application_logs_encryption.target_key_arn
  timeout      = 900
  memory       = 1024
  providers = {
    aws.region = aws.region
  }
}

# Additional IAM permissions
resource "aws_iam_role_policy" "s3_create_batch_replication_jobs" {
  name     = "create-s3-batch-replication-jobs-${data.aws_default_tags.current.tags.environment-name}"
  role     = var.create_s3_batch_replication_jobs_lambda_iam_role.id
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

resource "aws_scheduler_schedule" "invoke_lambda_every_15_minutes" {
  name = "invoke-lambda-every-2-minutes-${data.aws_default_tags.current.tags.environment-name}"

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = "rate(20 minutes)"
  state               = var.s3_replication.enable_s3_batch_job_replication_scheduler ? "ENABLED" : "DISABLED"

  target {
    arn      = module.s3_create_batch_replication_jobs.lambda.arn
    role_arn = aws_iam_role.scheduler_role.arn
  }
  provider = aws.region
}

# TODO: move this resource to global and pass it in using the roles variable
resource "aws_iam_role" "scheduler_role" {
  name               = "scheduler-${data.aws_default_tags.current.tags.environment-name}-${data.aws_region.current.name}"
  assume_role_policy = data.aws_iam_policy_document.scheduler_assume_role.json
  provider           = aws.region
}

data "aws_iam_policy_document" "scheduler_assume_role" {
  statement {
    actions = [
      "sts:AssumeRole",
    ]
    effect = "Allow"
    principals {
      type        = "Service"
      identifiers = ["scheduler.amazonaws.com"]
    }
  }
  provider = aws.region
}

resource "aws_iam_role_policy" "scheduler_invoke_lambda" {
  name     = "scheduler-invoke-lambda-${data.aws_default_tags.current.tags.environment-name}-${data.aws_region.current.name}"
  policy   = data.aws_iam_policy_document.scheduler_invoke_lambda.json
  role     = aws_iam_role.scheduler_role.id
  provider = aws.region
}

data "aws_iam_policy_document" "scheduler_invoke_lambda" {
  statement {
    sid    = "ScheduleInvokeLambda"
    effect = "Allow"
    resources = [
      module.s3_create_batch_replication_jobs.lambda.arn,
    ]
    actions = [
      "lambda:InvokeFunction",
    ]
  }
  provider = aws.region
}

resource "aws_iam_role_policy_attachment" "cloudwatch_lambda_insights" {
  role       = var.create_s3_batch_replication_jobs_lambda_iam_role.id
  policy_arn = "arn:aws:iam::aws:policy/CloudWatchLambdaInsightsExecutionRolePolicy"
  provider   = aws.region
}
