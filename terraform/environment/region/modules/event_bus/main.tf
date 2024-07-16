# Event bus for OPG events

resource "aws_cloudwatch_event_bus" "main" {
  name     = data.aws_default_tags.current.tags.environment-name
  provider = aws.region
}

resource "aws_cloudwatch_event_archive" "main" {
  name             = data.aws_default_tags.current.tags.environment-name
  event_source_arn = aws_cloudwatch_event_bus.main.arn
  provider         = aws.region
}

data "aws_kms_alias" "sqs" {
  name     = "alias/${data.aws_default_tags.current.tags.application}_sqs_secret_encryption_key"
  provider = aws.region
}

resource "aws_sqs_queue" "event_bus_dead_letter_queue" {
  name                              = "${data.aws_default_tags.current.tags.environment-name}-event-bus-dead-letter-queue"
  kms_master_key_id                 = data.aws_kms_alias.sqs.target_key_id
  kms_data_key_reuse_period_seconds = 300
  policy                            = data.aws_iam_policy_document.sqs.json
  provider                          = aws.region
}

data "aws_iam_policy_document" "sqs" {
  statement {
    sid    = "DeadLetterQueueAccess"
    effect = "Allow"
    principals {
      type        = "Service"
      identifiers = ["events.amazonaws.com"]
    }
    actions = [
      "sqs:SendMessage",
    ]
    condition {
      test     = "ArnEquals"
      variable = "aws:SourceArn"
      values = [
        aws_cloudwatch_event_rule.cross_account_put.arn
      ]
    }
  }
}

resource "aws_cloudwatch_metric_alarm" "event_bus_dead_letter_queue" {
  alarm_name          = "${data.aws_default_tags.current.tags.environment-name}-event-bus-dead-letter-queue"
  comparison_operator = "GreaterThanOrEqualToThreshold"
  evaluation_periods  = 1
  metric_name         = "ApproximateNumberOfMessagesVisible"
  namespace           = "AWS/SQS"
  period              = 60
  statistic           = "Sum"
  threshold           = 1
  alarm_description   = "${data.aws_default_tags.current.tags.environment-name} event bus dead letter queue has messages"
  alarm_actions       = [data.aws_sns_topic.custom_cloudwatch_alarms.arn]
  provider            = aws.region
}

# Send event to remote account event bus

resource "aws_iam_role_policy" "cross_account_put" {
  name     = "${data.aws_default_tags.current.tags.environment-name}-${data.aws_region.current.name}-cross-account-put"
  policy   = data.aws_iam_policy_document.cross_account_put_access.json
  role     = var.iam_role.id
  provider = aws.region
}

data "aws_iam_policy_document" "cross_account_put_access" {
  statement {
    sid    = "CrossAccountPutAccess"
    effect = "Allow"
    actions = [
      "events:PutEvents",
    ]
    resources = [
      var.target_event_bus_arn
    ]
  }
  provider = aws.region
}

resource "aws_cloudwatch_event_rule" "cross_account_put" {
  name           = "${data.aws_default_tags.current.tags.environment-name}-cross-account-put"
  description    = "forward events to bus in remote account"
  event_bus_name = aws_cloudwatch_event_bus.main.name

  event_pattern = jsonencode({
    source = ["opg.poas.makeregister"]
  })
  provider = aws.region
}

resource "aws_cloudwatch_event_target" "cross_account_put" {
  target_id      = "${data.aws_default_tags.current.tags.environment-name}-cross-account-put-event"
  event_bus_name = aws_cloudwatch_event_bus.main.name
  arn            = var.target_event_bus_arn
  dead_letter_config {
    arn = aws_sqs_queue.event_bus_dead_letter_queue.arn
  }
  rule     = aws_cloudwatch_event_rule.cross_account_put.name
  role_arn = var.iam_role.arn
  provider = aws.region
}

# Allow other accounts to send messages
data "aws_iam_policy_document" "cross_account_receive" {
  statement {
    sid    = "CrossAccountAccess"
    effect = "Allow"
    actions = [
      "events:PutEvents",
    ]
    resources = [
      aws_cloudwatch_event_bus.main.arn
    ]

    principals {
      type        = "AWS"
      identifiers = var.receive_account_ids
    }
  }
}

resource "aws_cloudwatch_event_bus_policy" "cross_account_receive" {
  count          = length(var.receive_account_ids) > 0 ? 1 : 0
  event_bus_name = aws_cloudwatch_event_bus.main.name
  policy         = data.aws_iam_policy_document.cross_account_receive.json
  provider       = aws.region
}
