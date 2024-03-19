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
