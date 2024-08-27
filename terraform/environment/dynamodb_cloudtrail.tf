data "aws_s3_bucket" "cloudtrail" {
  bucket   = "cloudtrail.${local.environment.account_name}.modernise.opg.service.justice.gov.uk"
  provider = aws.eu_west_1
}

resource "aws_cloudtrail" "dynamodb" {
  count                         = local.environment.dynamodb.cloudtrail_enabled ? 1 : 0
  name                          = "dynamodb"
  s3_bucket_name                = data.aws_s3_bucket.cloudtrail.id
  s3_key_prefix                 = "dynamodb"
  include_global_service_events = true
  is_multi_region_trail         = true
  event_selector {
    read_write_type = "All"
    # include_management_events = true

    data_resource {
      type = "AWS::DynamoDB::Table"
      values = [
        aws_dynamodb_table.lpas_table.arn,
        aws_dynamodb_table_replica.lpas_table[0].arn
      ]
    }
    data_resource {
      type   = "AWS::DynamoDB::Stream"
      values = ["${aws_dynamodb_table.lpas_table.arn}/stream/*"]
    }
  }
}
