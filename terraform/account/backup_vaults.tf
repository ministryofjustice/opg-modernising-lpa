resource "aws_backup_vault" "eu_west_1" {
  name     = "eu-west-1-${local.default_tags.account-name}-backup-vault"
  provider = aws.eu_west_1
}

resource "aws_backup_vault" "eu_west_2" {
  name     = "eu-west-2-${local.default_tags.account-name}-backup-vault"
  provider = aws.eu_west_2
}


data "aws_kms_alias" "sns_encryption_key_eu_west_1" {
  name     = "alias/${local.default_tags.application}_sns_secret_encryption_key"
  provider = aws.eu_west_1
}

resource "aws_sns_topic" "aws_backup_failure_events_eu_west_1" {
  name                                     = "${local.default_tags.account-name}-backup-vault-failure-events"
  kms_master_key_id                        = data.aws_kms_alias.sns_encryption_key_eu_west_1.target_key_arn
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
  provider                                 = aws.eu_west_1
}

data "aws_iam_policy_document" "aws_backup_sns_eu_west_1" {
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
  provider = aws.eu_west_1
}

resource "aws_sns_topic_policy" "aws_backup_failure_events_eu_west_1" {
  arn      = aws_sns_topic.aws_backup_failure_events.arn
  policy   = data.aws_iam_policy_document.aws_backup_sns.json
  provider = aws.eu_west_1
}

resource "aws_backup_vault_notifications" "aws_backup_failure_events_eu_west_1" {
  backup_vault_name   = data.aws_backup_vault.eu_west_1.name
  sns_topic_arn       = aws_sns_topic.aws_backup_failure_events.arn
  backup_vault_events = ["BACKUP_JOB_FAILED", "COPY_JOB_FAILED"]
  provider            = aws.eu_west_1
}

data "aws_kms_alias" "sns_encryption_key_eu_west_2" {
  name     = "alias/${local.default_tags.application}_sns_secret_encryption_key"
  provider = aws.eu_west_2
}

resource "aws_sns_topic" "aws_backup_failure_events_eu_west_2" {
  name                                     = "${local.default_tags.account-name}-backup-vault-failure-events"
  kms_master_key_id                        = data.aws_kms_alias.sns_encryption_key_eu_west_2.target_key_arn
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
  provider                                 = aws.eu_west_1
}

data "aws_iam_policy_document" "aws_backup_sns_eu_west_2" {
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
  provider = aws.eu_west_1
}

resource "aws_sns_topic_policy" "aws_backup_failure_events_eu_west_2" {
  arn      = aws_sns_topic.aws_backup_failure_events.arn
  policy   = data.aws_iam_policy_document.aws_backup_sns.json
  provider = aws.eu_west_1
}

resource "aws_backup_vault_notifications" "aws_backup_failure_events_eu_west_2" {
  backup_vault_name   = data.aws_backup_vault.eu_west_1.name
  sns_topic_arn       = aws_sns_topic.aws_backup_failure_events.arn
  backup_vault_events = ["BACKUP_JOB_FAILED", "COPY_JOB_FAILED"]
  provider            = aws.eu_west_1
}
