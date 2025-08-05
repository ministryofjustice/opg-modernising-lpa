module "schedule_runner" {
  source      = "../lambda"
  lambda_name = "schedule-runner"
  description = "Function to run scheduled tasks on a schedule set by EventBridge Scheduler"
  environment_variables = {
    EVENT_BUS_NAME                 = var.event_bus.name
    GOVUK_NOTIFY_BASE_URL          = "https://api.notifications.service.gov.uk"
    LPAS_TABLE                     = var.lpas_table.name
    SEARCH_ENDPOINT                = var.search_endpoint
    SEARCH_INDEX_NAME              = var.search_index_name
    SEARCH_INDEXING_DISABLED       = 1
    XRAY_ENABLED                   = 1
    LPA_STORE_BASE_URL             = var.lpa_store_base_url
    LPA_STORE_SECRET_ARN           = var.lpa_store_secret_arn
    APP_PUBLIC_URL                 = "https://${var.app_public_url}"
    CERTIFICATE_PROVIDER_START_URL = var.certificate_provider_start_url
    ATTORNEY_START_URL             = var.attorney_start_url
    ENVIRONMENT                    = data.aws_default_tags.current.tags.environment-name
  }
  image_uri            = "${var.lambda_function_image_ecr_url}:${var.lambda_function_image_tag}"
  aws_iam_role         = var.schedule_runner_lambda_role
  environment          = data.aws_default_tags.current.tags.environment-name
  kms_key              = data.aws_kms_alias.cloudwatch_application_logs_encryption.target_key_arn
  iam_policy_documents = [data.aws_iam_policy_document.schedule_runner.json]
  timeout              = 900
  memory               = 1024
  vpc_config = {
    subnet_ids         = var.vpc_config.subnet_ids
    security_group_ids = var.vpc_config.security_group_ids
  }
  providers = {
    aws.region = aws.region
  }
}

resource "aws_scheduler_schedule" "schedule_runner_hourly" {
  name = "schedule-runner-hourly-${data.aws_default_tags.current.tags.environment-name}"
  // TODO Revert before merging
  schedule_expression = "rate(1 minute)"
  description         = "Runs every hour"

  flexible_time_window {
    mode = "OFF"
  }

  target {
    arn      = module.schedule_runner.lambda.arn
    role_arn = var.schedule_runner_scheduler.arn
  }

  provider = aws.region
}

data "aws_kms_alias" "dynamodb_encryption_key" {
  name     = "alias/${data.aws_default_tags.current.tags.application}_dynamodb_encryption"
  provider = aws.region
}

data "aws_kms_alias" "secrets_manager_secret_encryption_key" {
  name     = "alias/${data.aws_default_tags.current.tags.application}_secrets_manager_secret_encryption_key"
  provider = aws.region
}

data "aws_kms_alias" "aws_lambda" {
  name     = "alias/aws/lambda"
  provider = aws.region
}

resource "aws_lambda_permission" "allow_cloudwatch_scheduler_to_call_schedule_runner" {
  statement_id   = "AllowExecutionFromCloudWatchMlpa"
  action         = "lambda:InvokeFunction"
  function_name  = module.schedule_runner.lambda.function_name
  principal      = "events.amazonaws.com"
  source_account = data.aws_caller_identity.current.account_id
  source_arn     = aws_scheduler_schedule.schedule_runner_hourly.arn
  provider       = aws.region
}

locals {
  policy_region_prefix = lower(replace(data.aws_region.current.name, "-", ""))
}
