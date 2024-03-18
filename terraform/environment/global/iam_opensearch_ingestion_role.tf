resource "aws_iam_role" "opensearch_ingestion" {
  name               = "${data.aws_default_tags.current.tags.environment-name}-opensearch-ingestion-role"
  assume_role_policy = data.aws_iam_policy_document.opensearch_ingestion.json
  provider           = aws.global
}

data "aws_iam_policy_document" "opensearch_ingestion" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      identifiers = ["osis-pipelines.amazonaws.com"]
      type        = "Service"
    }
  }
  provider = aws.global
}
