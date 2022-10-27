data "aws_backup_vault" "eu_west_1" {
  name     = "eu-west-1-${local.environment.account_name}-backup-vault"
  provider = aws.eu_west_1
}

data "aws_backup_vault" "eu_west_2" {
  name     = "eu-west-2-${local.environment.account_name}-backup-vault"
  provider = aws.eu_west_2
}

data "aws_iam_role" "aws_backup_role" {
  name     = "aws-backup-role"
  provider = aws.eu_west_1
}

resource "aws_backup_plan" "main" {
  count = local.environment.backups.backup_plan_enabled ? 1 : 0
  name  = "${local.environment_name}-main-backup-plan"

  rule {
    completion_window   = 10080
    recovery_point_tags = {}
    rule_name           = "DailyBackups"
    schedule            = "cron(0 5 ? * * *)"
    start_window        = 480
    target_vault_name   = data.aws_backup_vault.eu_west_1.name

    lifecycle {
      cold_storage_after = 0
      delete_after       = 90

    }
    dynamic "copy_action" {
      for_each = local.environment.backups.copy_action_enabled ? [1] : []
      content {
        destination_vault_arn = data.aws_backup_vault.eu_west_2.arn
      }
    }
  }
  rule {
    completion_window   = 10080
    recovery_point_tags = {}
    rule_name           = "Monthly"
    schedule            = "cron(0 5 1 * ? *)"
    start_window        = 480
    target_vault_name   = data.aws_backup_vault.eu_west_1.name

    lifecycle {
      cold_storage_after = 30
      delete_after       = 365
    }
    dynamic "copy_action" {
      for_each = local.environment.backups.copy_action_enabled ? [1] : []
      content {
        destination_vault_arn = data.aws_backup_vault.eu_west_2.arn
      }
    }
  }
  provider = aws.eu_west_1
}

resource "aws_backup_selection" "main" {
  count        = local.environment.backups.backup_plan_enabled ? 1 : 0
  iam_role_arn = data.aws_iam_role.aws_backup_role.arn
  name         = "${local.environment_name}_main_backup_selection"
  plan_id      = aws_backup_plan.main[0].id

  resources = [
    aws_dynamodb_table.lpas_table.arn,
  ]
  provider = aws.eu_west_1
}

data "aws_kms_alias" "sns_encryption_key_eu_west_1" {
  name     = "alias/${local.default_tags.application}_sns_secret_encryption_key"
  provider = aws.eu_west_1
}

resource "aws_sns_topic" "aws_backup_failure_events" {
  count                                    = local.environment.backups.backup_plan_enabled ? 1 : 0
  name                                     = "${local.environment_name}-backup-vault-failure-events"
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

data "aws_iam_policy_document" "aws_backup_sns" {
  count = local.environment.backups.backup_plan_enabled ? 1 : 0
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
      aws_sns_topic.aws_backup_failure_events[0].arn,
    ]
  }
  provider = aws.eu_west_1
}

resource "aws_sns_topic_policy" "aws_backup_failure_events" {
  count    = local.environment.backups.backup_plan_enabled ? 1 : 0
  arn      = aws_sns_topic.aws_backup_failure_events[0].arn
  policy   = data.aws_iam_policy_document.aws_backup_sns[0].json
  provider = aws.eu_west_1
}

resource "aws_backup_vault_notifications" "aws_backup_failure_events" {
  count               = local.environment.backups.backup_plan_enabled ? 1 : 0
  backup_vault_name   = data.aws_backup_vault.eu_west_1.name
  sns_topic_arn       = aws_sns_topic.aws_backup_failure_events[0].arn
  backup_vault_events = ["BACKUP_JOB_FAILED", "COPY_JOB_FAILED"]
  provider            = aws.eu_west_1
}
