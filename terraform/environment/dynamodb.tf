data "aws_kms_alias" "dynamodb_encryption_key_eu_west_1" {
  name     = "alias/${local.default_tags.application}_dynamodb_encryption"
  provider = aws.eu_west_1
}

data "aws_kms_alias" "dynamodb_encryption_key_eu_west_2" {
  name     = "alias/${local.default_tags.application}_dynamodb_encryption"
  provider = aws.eu_west_2
}

resource "aws_dynamodb_table" "lpas_table" {
  name                        = "${local.environment_name}-${local.environment.dynamodb.table_name}"
  billing_mode                = "PAY_PER_REQUEST"
  deletion_protection_enabled = local.default_tags.environment-name == "production" ? true : false
  # see docs/runbooks/disabling_dynamodb_global_tables.md when Global Tables needs to be disabled
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"
  hash_key         = "PK"
  range_key        = "SK"

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
