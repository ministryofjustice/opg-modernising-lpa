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
