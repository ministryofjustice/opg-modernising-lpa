# Event bus for opg.poas events

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
  provider                          = aws.region
}

resource "aws_sqs_queue_policy" "event_bus_dead_letter_queue_policy" {
  queue_url = aws_sqs_queue.event_bus_dead_letter_queue.id
  policy    = data.aws_iam_policy_document.event_bus_dead_letter_queue.json
  provider  = aws.region
}

data "aws_iam_policy_document" "event_bus_dead_letter_queue" {
  statement {
    sid    = "DeadLetterQueueAccess"
    effect = "Allow"
    principals {
      type        = "Service"
      identifiers = ["events.amazonaws.com"]
    }
    resources = [aws_sqs_queue.event_bus_dead_letter_queue.arn]
    actions   = ["sqs:SendMessage"]

    condition {
      test     = "ArnEquals"
      variable = "aws:SourceArn"
      values = [
        aws_cloudwatch_event_rule.cross_account_put.arn,
        "arn:aws:events:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:rule/${data.aws_default_tags.current.tags.environment-name}/${data.aws_default_tags.current.tags.environment-name}-receive-events-lpa-store",
        "arn:aws:events:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:rule/${data.aws_default_tags.current.tags.environment-name}/${data.aws_default_tags.current.tags.environment-name}-receive-events-mlpa",
        "arn:aws:events:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:rule/${data.aws_default_tags.current.tags.environment-name}/${data.aws_default_tags.current.tags.environment-name}-receive-events-sirius",
      ]
    }
  }
  provider = aws.region
}

resource "aws_sns_topic" "event_bus_dead_letter_queue" {
  name                                     = "${data.aws_default_tags.current.tags.environment-name}-event-bus-dead-letter-queue-alarms"
  kms_master_key_id                        = data.aws_kms_alias.sns_kms_key_alias.target_key_id
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
  alarm_actions       = [aws_sns_topic.event_bus_dead_letter_queue.arn]
  dimensions = {
    QueueName = aws_sqs_queue.event_bus_dead_letter_queue.name
  }
  provider = aws.region
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
    # replace "region" in each arn with the current region assuming that event busses in other accounts will exist in each region.
    resources = values({ for k, v in var.target_event_bus_arns : k => replace(v, "region", data.aws_region.current.name) })
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
  for_each       = var.target_event_bus_arns
  target_id      = "${data.aws_default_tags.current.tags.environment-name}-${each.key}-cross-account-put-event"
  event_bus_name = aws_cloudwatch_event_bus.main.name
  arn            = replace(each.value, "region", data.aws_region.current.name)
  dead_letter_config {
    arn = aws_sqs_queue.event_bus_dead_letter_queue.arn
  }
  rule     = aws_cloudwatch_event_rule.cross_account_put.name
  role_arn = var.iam_role.arn
  provider = aws.region
}

data "aws_iam_policy_document" "events_emitted" {
  statement {
    effect = "Allow"
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]
    resources = [
      "arn:aws:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:log-group:/aws/events/*:*"
    ]

    principals {
      type = "Service"
      identifiers = [
        "events.amazonaws.com",
        "delivery.logs.amazonaws.com",
      ]
    }
  }
}

resource "aws_cloudwatch_log_resource_policy" "events_emitted" {
  count           = var.log_emitted_events ? 1 : 0
  policy_document = data.aws_iam_policy_document.events_emitted.json
  policy_name     = "${data.aws_default_tags.current.tags.environment-name}-events-emitted"
  provider        = aws.region
}

#tfsec:ignore:aws-cloudwatch-log-group-customer-key
resource "aws_cloudwatch_log_group" "events_emitted" {
  count             = var.log_emitted_events ? 1 : 0
  name              = "/aws/events/${data.aws_default_tags.current.tags.environment-name}-emitted"
  retention_in_days = 5
  provider          = aws.region
}

resource "aws_cloudwatch_event_target" "events_emitted" {
  count          = var.log_emitted_events ? 1 : 0
  target_id      = "${data.aws_default_tags.current.tags.environment-name}-events-emitted-target"
  event_bus_name = aws_cloudwatch_event_bus.main.name
  rule           = aws_cloudwatch_event_rule.cross_account_put.name
  arn            = aws_cloudwatch_log_group.events_emitted[0].arn
  provider       = aws.region
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

resource "aws_cloudwatch_query_definition" "events_emitted" {
  count           = var.log_emitted_events ? 1 : 0
  name            = "${data.aws_default_tags.current.tags.environment-name}/emitted-events"
  log_group_names = [aws_cloudwatch_log_group.events_emitted[0].name]

  query_string = <<EOF
fields @timestamp, @message, @logStream, @log
| sort @timestamp desc
| limit 10000
EOF
  provider     = aws.region
}
