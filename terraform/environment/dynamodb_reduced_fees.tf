resource "aws_dynamodb_table" "reduced_fees" {
  name                        = "${local.environment_name}-reduced-fees"
  billing_mode                = "PAY_PER_REQUEST"
  deletion_protection_enabled = local.default_tags.environment-name == "production" ? true : false
  stream_enabled              = true
  stream_view_type            = "NEW_AND_OLD_IMAGES"

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

# resource "aws_dynamodb_table_replica" "reduced_fees" {
#   global_table_arn       = aws_dynamodb_table.reduced_fees.arn
#   kms_key_arn            = data.aws_kms_alias.dynamodb_encryption_key_eu_west_2.target_key_arn
#   point_in_time_recovery = true
#   provider               = aws.eu_west_2
# }

resource "aws_cloudwatch_event_bus" "reduced_fees" {
  name     = "reduced-fees"
  provider = aws.eu_west_1
}
