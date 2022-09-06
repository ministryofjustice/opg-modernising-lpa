data "aws_backup_vault" "eu-west-1" {
  name     = "eu-west-1-${local.environment.account_name}-backup-vault"
  provider = aws.eu_west_1
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
    target_vault_name   = data.aws_backup_vault.eu-west-1.name

    lifecycle {
      cold_storage_after = 0
      delete_after       = 90
    }
    # copy_action {
    #   destination_vault_arn = ""
    # }
  }
  rule {
    completion_window   = 10080
    recovery_point_tags = {}
    rule_name           = "Monthly"
    schedule            = "cron(0 5 1 * ? *)"
    start_window        = 480
    target_vault_name   = data.aws_backup_vault.eu-west-1.name

    lifecycle {
      cold_storage_after = 30
      delete_after       = 365
    }
    # copy_action {
    #   destination_vault_arn = ""
    # }
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

resource "aws_sns_topic" "aws_backup_failure_events" {
  count    = local.environment.backups.backup_plan_enabled ? 1 : 0
  name     = "backup-vault-failure-events"
  provider = aws.eu_west_1
}

data "aws_iam_policy_document" "aws_backup_sns" {
  policy_id = "__default_policy_ID"

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

    sid = "__default_statement_ID"
  }
}

resource "aws_sns_topic_policy" "aws_backup_failure_events" {
  count    = local.environment.backups.backup_plan_enabled ? 1 : 0
  arn      = aws_sns_topic.aws_backup_failure_events[0].arn
  policy   = data.aws_iam_policy_document.aws_backup_sns.json
  provider = aws.eu_west_1
}

resource "aws_backup_vault_notifications" "aws_backup_failure_events" {
  count               = local.environment.backups.backup_plan_enabled ? 1 : 0
  backup_vault_name   = data.aws_backup_vault.eu-west-1.name
  sns_topic_arn       = aws_sns_topic.aws_backup_failure_events[0].arn
  backup_vault_events = ["BACKUP_JOB_FAILED", "COPY_JOB_FAILED"]
  provider            = aws.eu_west_1
}
