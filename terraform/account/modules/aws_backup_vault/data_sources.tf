data "aws_region" "current" {
  provider = aws.region
}

data "aws_default_tags" "current" {
  provider = aws.region
}

data "aws_kms_alias" "sns_encryption_key" {
  name     = "alias/${data.aws_default_tags.current.tags.application}_sns_secret_encryption_key"
  provider = aws.region
}

data "aws_iam_role" "sns_success_feedback" {
  name     = "SNSSuccessFeedback"
  provider = aws.global
}

data "aws_iam_role" "sns_failure_feedback" {
  provider = aws.global
  name     = "SNSFailureFeedback"
}
