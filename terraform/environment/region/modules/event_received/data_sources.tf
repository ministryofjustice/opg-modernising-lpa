data "aws_default_tags" "current" {
  provider = aws.region
}

data "aws_region" "current" {
  provider = aws.region
}

data "aws_kms_alias" "cloudwatch_application_logs_encryption" {
  name     = "alias/${data.aws_default_tags.current.tags.application}_cloudwatch_application_logs_encryption"
  provider = aws.region
}

data "aws_kms_alias" "event_received_sqs" {
  name     = "alias/${data.aws_default_tags.current.tags.application}_event_received_sqs_secret_encryption_key"
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

data "aws_kms_alias" "jwt_key" {
  name     = "alias/opg-data-lpa-store/${data.aws_default_tags.current.tags.account-name}/jwt-key"
  provider = aws.management
}

data "aws_secretsmanager_secret" "lpa_store_jwt_key" {
  name     = "opg-data-lpa-store/${data.aws_default_tags.current.tags.account-name}/jwt-key"
  provider = aws.management
}

data "aws_kms_alias" "reduced_fees_uploads_s3_encryption" {
  name     = "alias/${data.aws_default_tags.current.tags.application}_reduced_fees_uploads_s3_encryption"
  provider = aws.region
}
