resource "aws_sns_topic" "s3_event_notification" {
  name                                     = "${replace(var.s3_bucket_id, ".", "-")}-s3-event-notification-topic"
  kms_master_key_id                        = aws_kms_key.s3_bucket_event_notification_sns.key_id
  application_failure_feedback_role_arn    = var.sns_failure_feedback_role_arn
  application_success_feedback_role_arn    = var.sns_success_feedback_role_arn
  application_success_feedback_sample_rate = 100
  firehose_failure_feedback_role_arn       = var.sns_failure_feedback_role_arn
  firehose_success_feedback_role_arn       = var.sns_success_feedback_role_arn
  firehose_success_feedback_sample_rate    = 100
  http_failure_feedback_role_arn           = var.sns_failure_feedback_role_arn
  http_success_feedback_role_arn           = var.sns_success_feedback_role_arn
  http_success_feedback_sample_rate        = 100
  lambda_failure_feedback_role_arn         = var.sns_failure_feedback_role_arn
  lambda_success_feedback_role_arn         = var.sns_success_feedback_role_arn
  lambda_success_feedback_sample_rate      = 100
  sqs_failure_feedback_role_arn            = var.sns_failure_feedback_role_arn
  sqs_success_feedback_role_arn            = var.sns_success_feedback_role_arn
  sqs_success_feedback_sample_rate         = 100
}

resource "aws_sns_topic_policy" "s3_event_notification" {
  arn    = aws_sns_topic.s3_event_notification.arn
  policy = data.aws_iam_policy_document.sns_topic_policy.json
}

data "aws_iam_policy_document" "sns_topic_policy" {
  statement {
    actions = [
      "SNS:Publish",
    ]
    condition {
      test     = "ArnLike"
      variable = "AWS:SourceArn"
      values   = ["arn:aws:s3:::${var.s3_bucket_id}"]
    }
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["s3.amazonaws.com"]
    }
    resources = [aws_sns_topic.s3_event_notification.arn]
  }
}
