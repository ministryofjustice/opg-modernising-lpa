# Event bus for reduced fees

resource "aws_cloudwatch_event_bus" "main" {
  name     = data.aws_default_tags.current.tags.environment-name
  provider = aws.region
}

resource "aws_cloudwatch_event_archive" "main" {
  name             = data.aws_default_tags.current.tags.environment-name
  event_source_arn = aws_cloudwatch_event_bus.main.arn
  provider         = aws.region
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
  name        = "${data.aws_default_tags.current.tags.environment-name}-cross-account-put"
  description = "forward events to bus in remote account"

  event_pattern = jsonencode({
    source = ["aws.dynamodb"]
  })
  provider = aws.region
}

resource "aws_cloudwatch_event_target" "cross_account_put" {
  target_id = "${data.aws_default_tags.current.tags.environment-name}-cross-account-put-event"
  arn       = var.target_event_bus_arn
  rule      = aws_cloudwatch_event_rule.cross_account_put.name
  role_arn  = var.iam_role.arn
  provider  = aws.region
}
