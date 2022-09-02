resource "aws_dynamodb_table" "lpas_tables" {
  name         = "${local.environment_name}-Lpas"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "Id"
  server_side_encryption {
    enabled = true
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
}
