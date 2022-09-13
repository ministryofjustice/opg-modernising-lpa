data "aws_kms_alias" "dynamodb_encryption_key_eu_west_1" {
  name     = "alias/${local.default_tags.application}_dynamodb_encryption"
  provider = aws.eu_west_1
}

resource "aws_dynamodb_table" "lpas_table" {
  name         = "${local.environment_name}-Lpas"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "Id"

  server_side_encryption {
    enabled     = true
    kms_key_arn = data.aws_kms_alias.dynamodb_encryption_key_eu_west_1.target_key_arn
  }

  attribute {
    name = "Id"
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
