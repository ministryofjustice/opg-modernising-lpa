# DynamoDB table for reduced fees

resource "aws_dynamodb_table" "reduced_fees" {
  name                        = "${data.aws_default_tags.current.tags.environment-name}-reduced-fees"
  billing_mode                = "PAY_PER_REQUEST"
  deletion_protection_enabled = data.aws_default_tags.current.tags.environment-name == "production" ? true : false
  stream_enabled              = true
  stream_view_type            = "NEW_AND_OLD_IMAGES"
  hash_key                    = "PK"

  # key for encryption may need to be available to consuming services if they intend to reach in and grab
  # server_side_encryption {
  #   enabled     = true
  #   kms_key_arn = var.dynamodb_encryption_key_arn
  # }

  attribute {
    name = "PK"
    type = "S"
  }

  point_in_time_recovery {
    enabled = true
  }

  lifecycle {
    ignore_changes = [replica]
  }
  provider = aws.region
}

# Event bus for reduced fees

resource "aws_cloudwatch_event_bus" "reduced_fees" {
  name     = "${data.aws_default_tags.current.tags.environment-name}-reduced-fees"
  provider = aws.region
}

resource "aws_cloudwatch_event_archive" "reduced_fees" {
  name             = "${data.aws_default_tags.current.tags.environment-name}-reduced-fees"
  event_source_arn = aws_cloudwatch_event_bus.reduced_fees.arn
  provider         = aws.region
}

# Event pipe to send events from dynamodb stream to event bus

resource "aws_pipes_pipe" "reduced_fees" {
  name          = "${data.aws_default_tags.current.tags.environment-name}-reduced-fees"
  description   = "capture events from dynamodb stream and pass to event bus"
  desired_state = "RUNNING"
  enrichment    = null
  name_prefix   = null
  role_arn      = aws_iam_role.reduced_fees_pipe.arn
  source        = aws_dynamodb_table.reduced_fees.stream_arn
  target        = aws_cloudwatch_event_bus.reduced_fees.arn
  source_parameters {
    dynamodb_stream_parameters {
      batch_size                         = 1
      maximum_batching_window_in_seconds = 0
      maximum_record_age_in_seconds      = -1
      maximum_retry_attempts             = 0
      on_partial_batch_item_failure      = null
      parallelization_factor             = 1
      starting_position                  = "LATEST"
    }
  }
  provider = aws.region
}

resource "aws_iam_role" "reduced_fees_pipe" {
  name               = "${data.aws_default_tags.current.tags.environment-name}-reduced-fees-pipe"
  assume_role_policy = data.aws_iam_policy_document.reduced_fees_assume_role.json
  path               = "/service-role/"
  provider           = aws.region
}

data "aws_iam_policy_document" "reduced_fees_assume_role" {
  statement {
    actions = ["sts:AssumeRole"]
    effect  = "Allow"
    principals {
      type        = "Service"
      identifiers = ["pipes.amazonaws.com"]
    }
    condition {
      test     = "StringEquals"
      variable = "aws:SourceAccount"
      values   = [data.aws_caller_identity.current.account_id]
    }
    condition {
      test     = "StringEquals"
      variable = "aws:SourceArn"
      values   = ["arn:aws:pipes:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:pipe/${data.aws_default_tags.current.tags.environment-name}-reduced-fees"]
    }
  }
  provider = aws.region
}

resource "aws_iam_role_policy" "reduced_fees_pipe_source" {
  name     = "${data.aws_default_tags.current.tags.environment-name}-DynamoDbPipeSource"
  policy   = data.aws_iam_policy_document.reduced_fees_dynamodb_source.json
  role     = aws_iam_role.reduced_fees_pipe.id
  provider = aws.region
}

data "aws_iam_policy_document" "reduced_fees_dynamodb_source" {
  statement {
    actions = [
      "dynamodb:DescribeStream",
      "dynamodb:GetRecords",
      "dynamodb:GetShardIterator",
      "dynamodb:ListStreams",
    ]
    effect    = "Allow"
    resources = [aws_dynamodb_table.reduced_fees.stream_arn]
  }
  provider = aws.region
}

resource "aws_iam_role_policy" "reduced_fees_pipe_target" {
  name     = "${data.aws_default_tags.current.tags.environment-name}-EventBusPipeTarget"
  policy   = data.aws_iam_policy_document.reduced_fees_eventbus_target.json
  role     = aws_iam_role.reduced_fees_pipe.id
  provider = aws.global
}

data "aws_iam_policy_document" "reduced_fees_eventbus_target" {
  statement {
    actions = [
      "events:PutEvents"
    ]
    effect    = "Allow"
    resources = [aws_cloudwatch_event_bus.reduced_fees.arn]
  }
  provider = aws.region
}

# Send event to remote account event bus

resource "aws_iam_role" "cross_account_put" {
  name               = "${data.aws_default_tags.current.tags.environment-name}-cross-account-put"
  assume_role_policy = data.aws_iam_policy_document.cross_account_put_assume_role.json
  provider           = aws.region
}

resource "aws_iam_role_policy" "cross_account_put" {
  name     = "${data.aws_default_tags.current.tags.environment-name}-cross-account-put"
  policy   = data.aws_iam_policy_document.cross_account_put_access.json
  role     = aws_iam_role.cross_account_put.id
  provider = aws.region
}

data "aws_iam_policy_document" "cross_account_put_assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["events.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
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
  description = "forward dynamodb stream events to bus in remote account"

  event_pattern = jsonencode({
    source = ["aws.dynamodb"]
  })
  provider = aws.region
}

resource "aws_cloudwatch_event_target" "cross_account_put" {
  target_id = "${data.aws_default_tags.current.tags.environment-name}-cross-account-put-event"
  arn       = var.target_event_bus_arn
  rule      = aws_cloudwatch_event_rule.cross_account_put.name
  role_arn  = aws_iam_role.cross_account_put.arn
  provider  = aws.region
}
