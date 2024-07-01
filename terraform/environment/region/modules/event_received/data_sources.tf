data "aws_default_tags" "current" {
  provider = aws.region
}

data "aws_region" "current" {
  provider = aws.region
}

data "aws_caller_identity" "current" {
  provider = aws.region
}

data "aws_kms_alias" "cloudwatch_application_logs_encryption" {
  name     = "alias/${data.aws_default_tags.current.tags.application}_cloudwatch_application_logs_encryption"
  provider = aws.region
}

data "aws_secretsmanager_secret" "gov_uk_notify_api_key" {
  name     = "gov-uk-notify-api-key"
  provider = aws.region
}

data "aws_secretsmanager_secret" "lpa_store_jwt_secret_key" {
  name     = "lpa-store-jwt-secret-key"
  provider = aws.region
}
