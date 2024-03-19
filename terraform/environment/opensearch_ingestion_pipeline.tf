resource "aws_iam_role_policy" "opensearch_ingestion" {
  name     = "opensearch_ingestion"
  role     = module.global.iam_roles.opensearch_ingestion.name
  policy   = data.aws_iam_policy_document.opensearch_ingestion.json
  provider = aws.global
}

data "aws_iam_policy_document" "opensearch_ingestion" {
  version = "2012-10-17"

  statement {
    sid    = "allowRunExportJob"
    effect = "Allow"
    actions = [
      "dynamodb:DescribeTable",
      "dynamodb:DescribeContinuousBackups",
      "dynamodb:ExportTableToPointInTime",
    ]
    resources = [
      aws_dynamodb_table.lpas_table.arn,
    ]
  }

  statement {
    sid    = "allowCheckExportjob"
    effect = "Allow"
    actions = [
      "dynamodb:DescribeExport",
    ]
    resources = [
      "${aws_dynamodb_table.lpas_table.arn}/export/*",
    ]
  }

  statement {
    sid    = "allowReadFromStream"
    effect = "Allow"
    actions = [
      "dynamodb:DescribeStream",
      "dynamodb:GetRecords",
      "dynamodb:GetShardIterator",
    ]
    resources = [
      aws_dynamodb_table.lpas_table.stream_arn,
    ]
  }

  statement {
    sid    = "allowReadAndWriteToS3ForExport"
    effect = "Allow"
    actions = [
      "s3:GetObject",
      "s3:AbortMultipartUpload",
      "s3:PutObject",
      "s3:PutObjectAcl",
    ]
    resources = [
      "${aws_s3_bucket.opensearch_ingestion.arn}/export/*",
    ]
  }
}

data "aws_vpc" "main" {
  filter {
    name   = "tag:application"
    values = [local.default_tags.application]
  }
  provider = aws.eu_west_1
}

resource "aws_security_group" "opensearch_ingestion" {
  name_prefix = "${local.default_tags.environment-name}-opensearch-ingestion"
  vpc_id      = data.aws_vpc.main.id
  provider    = aws.eu_west_1
}

resource "aws_cloudwatch_log_group" "opensearch_ingestion" {
  name              = "/aws/vendedlogs/OpenSearchIngestion/lpas-${local.default_tags.environment-name}/audit-logs"
  retention_in_days = 1
  provider          = aws.eu_west_1
}

resource "aws_s3_bucket" "opensearch_ingestion" {
  bucket   = "${local.default_tags.environment-name}-opensearch-ingestion"
  provider = aws.eu_west_1
}

resource "aws_s3_bucket_acl" "opensearch_ingestion" {
  bucket   = aws_s3_bucket.opensearch_ingestion.bucket
  acl      = "private"
  provider = aws.eu_west_1
}
