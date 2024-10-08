data "aws_iam_policy_document" "event_bus_dead_letter_queue_policy" {
  statement {
    sid    = "First"
    effect = "Allow"

    principals {
      type        = "*"
      identifiers = ["*"]
    }

    actions   = ["sqs:SendMessage"]
    resources = [var.event_bus_dead_letter_queue.arn]

    condition {
      test     = "ArnEquals"
      variable = "aws:SourceArn"
      values = [
        aws_cloudwatch_event_rule.receive_events_lpa_store.arn,
        aws_cloudwatch_event_rule.receive_events_mlpa.arn,
        aws_cloudwatch_event_rule.receive_events_sirius.arn,
      ]
    }
  }
  provider = aws.region
}

resource "aws_sqs_queue_policy" "event_bus_dead_letter_queue_policy" {
  queue_url = var.event_bus_dead_letter_queue.id
  policy    = data.aws_iam_policy_document.event_bus_dead_letter_queue_policy.json
  provider  = aws.region
}
