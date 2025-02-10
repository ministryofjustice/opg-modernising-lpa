data "aws_sns_topic" "custom_cloudwatch_alarms" {
  name     = "custom_cloudwatch_alarms"
  provider = aws.region
}

data "aws_region" "current" {
  provider = aws.region
}

data "aws_default_tags" "current" {
  provider = aws.region
}

data "aws_caller_identity" "current" {
  provider = aws.region
}

data "aws_iam_role" "sns_success_feedback" {
  name     = "SNSSuccessFeedback"
  provider = aws.global
}

data "aws_iam_role" "sns_failure_feedback" {
  name     = "SNSFailureFeedback"
  provider = aws.global
}

data "aws_kms_alias" "sns_kms_key_alias" {
  name     = "alias/${data.aws_default_tags.current.tags.application}_sns_secret_encryption_key"
  provider = aws.region
}
