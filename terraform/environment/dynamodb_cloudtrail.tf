data "aws_s3_bucket" "cloudtrail" {
  bucket   = "cloudtrail.${local.environment.account_name}.modernise.opg.service.justice.gov.uk"
  provider = aws.eu_west_1
}

data "aws_kms_alias" "cloudtrail" {
  name     = "alias/cloudtrail_s3_modernising_lpa_${local.environment.account_name}"
  provider = aws.eu_west_1
}

data "aws_kms_alias" "dynamodb_cloudtrail_log_group" {
  name     = "alias/${local.default_tags.application}-dynamodb-cloudtrail-log-group-encryption"
  provider = aws.eu_west_1
}

resource "aws_cloudwatch_log_group" "cloudtrail_dynamodb" {
  count             = local.environment.dynamodb.cloudtrail_enabled ? 1 : 0
  name              = "/aws/cloudtrail/dynamodb-${local.default_tags.environment-name}"
  retention_in_days = 365
  kms_key_id        = data.aws_kms_alias.dynamodb_cloudtrail_log_group.target_key_arn
  provider          = aws.eu_west_1
}

data "aws_iam_role" "cloudtrail" {
  name     = "modernising-lpa-${local.environment.account_name}"
  provider = aws.eu_west_1
}

resource "aws_cloudtrail" "dynamodb" {
  count                         = local.environment.dynamodb.cloudtrail_enabled ? 1 : 0
  name                          = "${local.default_tags.environment-name}-dynamodb"
  s3_bucket_name                = data.aws_s3_bucket.cloudtrail.id
  kms_key_id                    = data.aws_kms_alias.cloudtrail.target_key_arn
  cloud_watch_logs_group_arn    = "${aws_cloudwatch_log_group.cloudtrail_dynamodb[0].arn}:*"
  cloud_watch_logs_role_arn     = data.aws_iam_role.cloudtrail.arn
  s3_key_prefix                 = "${local.default_tags.environment-name}-dynamodb"
  enable_log_file_validation    = true
  include_global_service_events = true
  is_multi_region_trail         = true
  event_selector {
    read_write_type           = "All"
    include_management_events = false

    data_resource {
      type = "AWS::DynamoDB::Table"
      values = [
        aws_dynamodb_table.lpas_table.arn,
        local.environment.dynamodb.region_replica_enabled ? aws_dynamodb_table_replica.lpas_table[0].arn : ""
      ]
    }
  }
  provider = aws.eu_west_1
}

resource "aws_cloudwatch_query_definition" "dynamodb_cloudtrail_table" {
  count           = local.environment.dynamodb.cloudtrail_enabled ? 1 : 0
  name            = "${local.default_tags.environment-name}/dynamodb cloudtrail table"
  log_group_names = [aws_cloudwatch_log_group.cloudtrail_dynamodb[0].name]

  query_string = <<EOF
fields @timestamp, eventSource, eventName, errorCode, userIdentity.principalId, userIdentity.sessionContext.sessionIssuer.userName
| filter resources.1.type != "AWS::DynamoDB::Stream"
| sort @timestamp desc
EOF
  provider     = aws.eu_west_1
}

resource "aws_cloudwatch_query_definition" "dynamodb_cloudtrail_stream" {
  count           = local.environment.dynamodb.cloudtrail_enabled ? 1 : 0
  name            = "${local.default_tags.environment-name}/dynamodb cloudtrail stream"
  log_group_names = [aws_cloudwatch_log_group.cloudtrail_dynamodb[0].name]

  query_string = <<EOF
fields @timestamp, eventSource, eventName, errorCode, userIdentity.principalId, userIdentity.sessionContext.sessionIssuer.userName
| filter resources.1.type == "AWS::DynamoDB::Stream"
| sort @timestamp desc
EOF
  provider     = aws.eu_west_1
}
