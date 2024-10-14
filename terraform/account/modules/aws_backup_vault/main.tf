resource "aws_backup_vault" "main" {
  name     = "${data.aws_region.current.name}-${data.aws_default_tags.current.tags.account-name}-backup-vault"
  provider = aws.region
}

resource "aws_sns_topic" "aws_backup_failure_events" {
  name                                     = "${data.aws_default_tags.current.tags.account-name}-backup-vault-failure-events"
  kms_master_key_id                        = data.aws_kms_alias.sns_encryption_key.target_key_arn
  application_failure_feedback_role_arn    = data.aws_iam_role.sns_failure_feedback.arn
  application_success_feedback_role_arn    = data.aws_iam_role.sns_success_feedback.arn
  application_success_feedback_sample_rate = 100
  firehose_failure_feedback_role_arn       = data.aws_iam_role.sns_failure_feedback.arn
  firehose_success_feedback_role_arn       = data.aws_iam_role.sns_success_feedback.arn
  firehose_success_feedback_sample_rate    = 100
  http_failure_feedback_role_arn           = data.aws_iam_role.sns_failure_feedback.arn
  http_success_feedback_role_arn           = data.aws_iam_role.sns_success_feedback.arn
  http_success_feedback_sample_rate        = 100
  lambda_failure_feedback_role_arn         = data.aws_iam_role.sns_failure_feedback.arn
  lambda_success_feedback_role_arn         = data.aws_iam_role.sns_success_feedback.arn
  lambda_success_feedback_sample_rate      = 100
  sqs_failure_feedback_role_arn            = data.aws_iam_role.sns_failure_feedback.arn
  sqs_success_feedback_role_arn            = data.aws_iam_role.sns_success_feedback.arn
  sqs_success_feedback_sample_rate         = 100
  provider                                 = aws.region
}

data "aws_iam_policy_document" "aws_backup_sns" {
  statement {
    actions = [
      "SNS:Publish",
    ]

    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["backup.amazonaws.com"]
    }

    resources = [
      aws_sns_topic.aws_backup_failure_events.arn,
    ]
  }
  provider = aws.region
}

resource "aws_sns_topic_policy" "aws_backup_failure_events" {
  arn      = aws_sns_topic.aws_backup_failure_events.arn
  policy   = data.aws_iam_policy_document.aws_backup_sns.json
  provider = aws.region
}

resource "aws_backup_vault_notifications" "aws_backup_failure_events" {
  backup_vault_name   = aws_backup_vault.main.name
  sns_topic_arn       = aws_sns_topic.aws_backup_failure_events.arn
  backup_vault_events = ["BACKUP_JOB_FAILED", "COPY_JOB_FAILED"]
  provider            = aws.region
}
