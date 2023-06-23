data "aws_caller_identity" "main" {
  provider = aws.global
}

data "aws_region" "eu_west_1" {
  provider = aws.eu_west_1
}

# DynamoDB table for reduced fees

resource "aws_dynamodb_table" "reduced_fees" {
  name                        = "${local.environment_name}-reduced-fees"
  billing_mode                = "PAY_PER_REQUEST"
  deletion_protection_enabled = local.default_tags.environment-name == "production" ? true : false
  stream_enabled              = true
  stream_view_type            = "NEW_AND_OLD_IMAGES"
  hash_key                    = "PK"

  # key for encryption may need to be available to consuming services if they intend to reach in and grab
  # server_side_encryption {
  #   enabled     = true
  #   kms_key_arn = data.aws_kms_alias.dynamodb_encryption_key_eu_west_1.target_key_arn
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
  provider = aws.eu_west_1
}

resource "aws_dynamodb_table_replica" "reduced_fees" {
  global_table_arn = aws_dynamodb_table.reduced_fees.arn
  # kms_key_arn            = data.aws_kms_alias.dynamodb_encryption_key_eu_west_2.target_key_arn
  point_in_time_recovery = true
  provider               = aws.eu_west_2
}

# Event bus for reduced fees

resource "aws_cloudwatch_event_bus" "reduced_fees" {
  name     = "reduced-fees"
  provider = aws.eu_west_1
}

resource "aws_cloudwatch_event_archive" "reduced_fees" {
  name             = "reduced-fees"
  event_source_arn = aws_cloudwatch_event_bus.reduced_fees.arn
  provider         = aws.eu_west_1
}

# Event pipe to send events from dynamodb stream to event bus

resource "aws_pipes_pipe" "reduced_fees" {
  name        = "reduced-fees"
  description = "capture events from dynamodb stream and pass to event bus"
  role_arn    = aws_iam_role.reduced_fees_pipe.arn
  source      = aws_dynamodb_table.reduced_fees.stream_arn
  target      = aws_cloudwatch_event_bus.reduced_fees.arn

  source_parameters {}
  target_parameters {}
  provider = aws.eu_west_1
}


resource "aws_iam_role" "reduced_fees_pipe" {
  assume_role_policy = data.aws_iam_policy_document.reduced_fees_assume_role.json
  path               = "/service-role/"
  managed_policy_arns = [
    "arn:aws:iam::653761790766:policy/service-role/DynamoDbPipeSourceTemplate-d47c4614",
    "arn:aws:iam::653761790766:policy/service-role/EventBusPipeTargetTemplate-102dc19b",
  ]
  provider = aws.global
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
      values   = [data.aws_caller_identity.main.account_id]
    }
    condition {
      test     = "StringEquals"
      variable = "aws:SourceArn"
      values   = ["arn:aws:pipes:${data.aws_region.eu_west_1.name}:${data.aws_caller_identity.main.account_id}:pipe/reduced-fees"]
    }
  }
  provider = aws.global
}

resource "aws_iam_role_policy" "reduced_fees_pipe_source" {
  name     = "${local.default_tags.environment-name}-DynamoDbPipeSource"
  policy   = data.aws_iam_policy_document.reduced_fees_dynamodb_source.json
  role     = aws_iam_role.reduced_fees_pipe.id
  provider = aws.global
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
  provider = aws.global
}

resource "aws_iam_role_policy" "reduced_fees_pipe_target" {
  name     = "${local.default_tags.environment-name}-EventBusPipeTarget"
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
  provider = aws.global
}

# Send event to remote account event bus

resource "aws_iam_role" "cross_account_put" {
  assume_role_policy = data.aws_iam_policy_document.cross_account_put_assume_role.json
  provider           = aws.global
}

resource "aws_iam_role_policy" "cross_account_put" {
  name     = "${local.default_tags.environment-name}-cross-account-put"
  policy   = data.aws_iam_policy_document.cross_account_put_access.json
  role     = aws_iam_role.cross_account_put.id
  provider = aws.global
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
}

data "aws_iam_policy_document" "cross_account_put_access" {
  statement {
    sid    = "CrossAccountPutAccess"
    effect = "Allow"
    actions = [
      "events:PutEvents",
    ]
    resources = [
      "arn:aws:events:eu-west-1:123456789012:event-bus/default"
    ]

    principals {
      type        = "AWS"
      identifiers = [data.aws_caller_identity.main.account_id]
    }
  }
  provider = aws.eu_west_1
}

# resource "aws_cloudwatch_event_bus_policy" "cross_account_put_access" {
#   policy         = data.aws_iam_policy_document.cross_account_put_access.json
#   event_bus_name = aws_cloudwatch_event_bus.reduced_fees.name
#   provider       = aws.eu_west_1
# }

resource "aws_cloudwatch_event_rule" "cross_account_put" {
  name        = "cross-account-put"
  description = "forward dynamodb stream events to bus in remote account"

  # event_pattern = jsonencode({
  #   detail-type = [
  #     "AWS Console Sign In via CloudTrail"
  #   ]
  # })
}

resource "aws_cloudwatch_event_target" "cross_account_put" {
  target_id = "CrossAccountPutEvent"
  arn       = "arn:aws:events:eu-west-1:123456789012:event-bus/default"
  rule      = aws_cloudwatch_event_rule.cross_account_put.name
  role_arn  = aws_iam_role.cross_account_put.arn
}
