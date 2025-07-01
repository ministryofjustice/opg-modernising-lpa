data "aws_kms_alias" "dynamodb_encryption_key_eu_west_1" {
  name     = "alias/${local.default_tags.application}_dynamodb_encryption"
  provider = aws.eu_west_1
}

data "aws_kms_alias" "dynamodb_encryption_key_eu_west_2" {
  name     = "alias/${local.default_tags.application}_dynamodb_encryption"
  provider = aws.eu_west_2
}

resource "aws_dynamodb_table" "lpas_table" {
  name                        = "${local.environment_name}-${local.environment.dynamodb.lpas_table_name}"
  billing_mode                = "PAY_PER_REQUEST"
  deletion_protection_enabled = local.default_tags.environment-name == "production" ? true : false
  stream_enabled              = true
  stream_view_type            = "NEW_AND_OLD_IMAGES"
  hash_key                    = "PK"
  range_key                   = "SK"

  global_secondary_index {
    name            = "SKUpdatedAtIndex"
    hash_key        = "SK"
    range_key       = "UpdatedAt"
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "LpaUIDIndex"
    hash_key        = "LpaUID"
    projection_type = "KEYS_ONLY"
  }

  server_side_encryption {
    enabled     = true
    kms_key_arn = data.aws_kms_alias.dynamodb_encryption_key_eu_west_1.target_key_arn
  }

  attribute {
    name = "PK"
    type = "S"
  }

  attribute {
    name = "SK"
    type = "S"
  }

  attribute {
    name = "LpaUID"
    type = "S"
  }

  attribute {
    name = "UpdatedAt"
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

resource "aws_dynamodb_table_replica" "lpas_table" {
  count                  = local.environment.dynamodb.region_replica_enabled ? 1 : 0
  global_table_arn       = aws_dynamodb_table.lpas_table.arn
  kms_key_arn            = data.aws_kms_alias.dynamodb_encryption_key_eu_west_2.target_key_arn
  point_in_time_recovery = true
  provider               = aws.eu_west_2
}

resource "aws_dynamodb_resource_policy" "lpas_table_replica" {
  count        = local.environment.dynamodb.region_replica_enabled ? 1 : 0
  resource_arn = aws_dynamodb_table_replica.lpas_table[0].arn
  policy       = data.aws_iam_policy_document.lpas_table.json
  provider     = aws.eu_west_2
}

resource "aws_dynamodb_resource_policy" "lpas_table" {
  resource_arn = aws_dynamodb_table.lpas_table.arn
  policy       = data.aws_iam_policy_document.lpas_table.json
  provider     = aws.eu_west_1
}

data "aws_iam_policy_document" "lpas_table" {
  statement {
    sid    = "AllowAccessForAppAndEventsReceived"
    effect = "Allow"
    actions = [
      "dynamodb:BatchGetItem",
      "dynamodb:DeleteItem",
      "dynamodb:GetItem",
      "dynamodb:PutItem",
      "dynamodb:Query",
      "dynamodb:Scan",
      "dynamodb:UpdateItem",
    ]
    resources = ["*"]

    principals {
      type = "AWS"
      identifiers = [
        module.global.iam_roles.app_ecs_task_role.arn,
        module.global.iam_roles.event_received_lambda.arn,
        data.aws_iam_role.aws_backup_role.arn,
      ]
    }
  }

  statement {
    sid    = "AllowAccessForScheduleRunner"
    effect = "Allow"
    actions = [
      "dynamodb:DeleteItem",
      "dynamodb:GetItem",
      "dynamodb:PutItem",
      "dynamodb:Query",
      "dynamodb:UpdateItem",
    ]
    resources = ["*"]

    principals {
      type = "AWS"
      identifiers = [
        module.global.iam_roles.schedule_runner_lambda.arn,
      ]
    }
  }

  statement {
    sid    = "AllowAccessForOpensearchPipeline"
    effect = "Allow"
    actions = [
      "dynamodb:DescribeTable",
      "dynamodb:DescribeContinuousBackups",
      "dynamodb:ExportTableToPointInTime",
    ]
    resources = ["*"]

    principals {
      type = "AWS"
      identifiers = [
        module.global.iam_roles.opensearch_pipeline.arn,
      ]
    }
  }

  statement {
    sid    = "DescribeExports"
    effect = "Allow"
    actions = [
      "dynamodb:DescribeExport",
    ]
    resources = [
      "*",
    ]
    principals {
      type = "AWS"
      identifiers = [
        module.global.iam_roles.opensearch_pipeline.arn,
      ]
    }
  }

  statement {
    sid    = "AllowReadAccessForUserRoles"
    effect = "Allow"
    actions = [
      "dynamodb:BatchGetItem",
      "dynamodb:GetItem",
      "dynamodb:Query",
      "dynamodb:Scan",
    ]
    resources = ["*"]

    principals {
      type = "AWS"
      identifiers = [
        local.environment.account_name == "development" ? "arn:aws:iam::${data.aws_caller_identity.global.account_id}:role/operator" : "arn:aws:iam::${data.aws_caller_identity.global.account_id}:role/data-access",
      ]
    }
  }

  statement {
    sid    = "AllowAccessForBreakglass"
    effect = "Allow"
    actions = [
      "dynamodb:*",
    ]
    resources = ["*"]

    principals {
      type = "AWS"
      identifiers = [
        "arn:aws:iam::${data.aws_caller_identity.global.account_id}:role/breakglass"
      ]
    }
  }
  provider = aws.global
}

resource "aws_dynamodb_table" "sessions_table" {
  name                        = "${local.environment_name}-${local.environment.dynamodb.sessions_table_name}"
  billing_mode                = "PAY_PER_REQUEST"
  deletion_protection_enabled = local.default_tags.environment-name == "production" ? true : false
  hash_key                    = "PK"
  range_key                   = "SK"

  server_side_encryption {
    enabled     = true
    kms_key_arn = data.aws_kms_alias.dynamodb_encryption_key_eu_west_1.target_key_arn
  }

  attribute {
    name = "PK"
    type = "S"
  }

  attribute {
    name = "SK"
    type = "S"
  }

  ttl {
    attribute_name = "ExpiresAt"
    enabled        = true
  }

  lifecycle {
    ignore_changes = [replica]
  }
  provider = aws.eu_west_1
}

resource "aws_dynamodb_table_replica" "sessions_table" {
  count                  = local.environment.dynamodb.region_replica_enabled ? 1 : 0
  global_table_arn       = aws_dynamodb_table.sessions_table.arn
  kms_key_arn            = data.aws_kms_alias.dynamodb_encryption_key_eu_west_2.target_key_arn
  point_in_time_recovery = true
  provider               = aws.eu_west_2
}

resource "aws_dynamodb_resource_policy" "sessions_table_replica" {
  count        = local.environment.dynamodb.region_replica_enabled ? 1 : 0
  resource_arn = aws_dynamodb_table_replica.sessions_table[0].arn
  policy       = data.aws_iam_policy_document.sessions_table.json
  provider     = aws.eu_west_2
}

resource "aws_dynamodb_resource_policy" "sessions_table" {
  resource_arn = aws_dynamodb_table.sessions_table.arn
  policy       = data.aws_iam_policy_document.sessions_table.json
  provider     = aws.eu_west_1
}

data "aws_iam_policy_document" "sessions_table" {
  statement {
    sid    = "AllowAccessForApp"
    effect = "Allow"
    actions = [
      "dynamodb:BatchGetItem",
      "dynamodb:DeleteItem",
      "dynamodb:GetItem",
      "dynamodb:PutItem",
      "dynamodb:Query",
      "dynamodb:Scan",
      "dynamodb:UpdateItem",
    ]
    resources = ["*"]

    principals {
      type = "AWS"
      identifiers = [
        module.global.iam_roles.app_ecs_task_role.arn,
        data.aws_iam_role.aws_backup_role.arn,
      ]
    }
  }

  statement {
    sid    = "AllowReadAccessForUserRoles"
    effect = "Allow"
    actions = [
      "dynamodb:BatchGetItem",
      "dynamodb:GetItem",
      "dynamodb:Query",
      "dynamodb:Scan",
    ]
    resources = ["*"]

    principals {
      type = "AWS"
      identifiers = [
        local.environment.account_name == "development" ? "arn:aws:iam::${data.aws_caller_identity.global.account_id}:role/operator" : "arn:aws:iam::${data.aws_caller_identity.global.account_id}:role/data-access",
      ]
    }
  }

  statement {
    sid    = "AllowAccessForBreakglass"
    effect = "Allow"
    actions = [
      "dynamodb:*",
    ]
    resources = ["*"]

    principals {
      type = "AWS"
      identifiers = [
        "arn:aws:iam::${data.aws_caller_identity.global.account_id}:role/breakglass"
      ]
    }
  }
  provider = aws.global
}
