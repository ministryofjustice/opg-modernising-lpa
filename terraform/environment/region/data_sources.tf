data "aws_kms_alias" "sns_kms_key_alias_global" {
  name     = "alias/${data.aws_default_tags.current.tags.application}_sns_secret_encryption_key"
  provider = aws.global
}

data "aws_iam_role" "sns_success_feedback" {
  name     = "SNSSuccessFeedback"
  provider = aws.global
}

data "aws_iam_role" "sns_failure_feedback" {
  name     = "SNSFailureFeedback"
  provider = aws.global
}
