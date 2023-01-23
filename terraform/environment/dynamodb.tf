data "aws_kms_alias" "dynamodb_encryption_key_eu_west_1" {
  name     = "alias/${local.default_tags.application}_dynamodb_encryption"
  provider = aws.eu_west_1
}

data "aws_kms_alias" "dynamodb_encryption_key_eu_west_2" {
  name     = "alias/${local.default_tags.application}_dynamodb_encryption"
  provider = aws.eu_west_2
}

resource "aws_dynamodb_table" "lpas_table" {
  name         = "${local.environment_name}-Lpas"
  billing_mode = "PAY_PER_REQUEST"
  # see docs/runbooks/disabling_dynamodb_global_tables.md when Global Tables needs to be disabled
  stream_enabled   = local.environment.dynamodb.stream_enabled
  stream_view_type = local.environment.dynamodb.stream_enabled ? "NEW_AND_OLD_IMAGES" : null
  hash_key         = "PK"
  range_key        = "SK"

  server_side_encryption {
    enabled     = true
    kms_key_arn = data.aws_kms_alias.dynamodb_encryption_key_eu_west_1.target_key_arn
  }

  dynamic "replica" {
    for_each = local.environment.dynamodb.region_replica_enabled ? [1] : []
    content {
      region_name            = "eu-west-2"
      kms_key_arn            = data.aws_kms_alias.dynamodb_encryption_key_eu_west_2.target_key_arn
      point_in_time_recovery = true
      propagate_tags         = true
    }
  }

  attribute {
    name = "PK"
    type = "S"
  }

  attribute {
    name = "SK"
    type = "S"
  }

  point_in_time_recovery {
    enabled = true
  }

  lifecycle {
    prevent_destroy = false
  }

  provider = aws.eu_west_1
}
