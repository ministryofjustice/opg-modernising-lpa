resource "aws_iam_role" "opensearch_pipeline" {
  name               = "${data.aws_default_tags.current.tags.environment-name}-opensearch-pipeline-role"
  assume_role_policy = data.aws_iam_policy_document.opensearch_pipeline.json
  provider           = aws.global
}

data "aws_iam_policy_document" "opensearch_pipeline" {
  # statement {
  #   effect  = "Allow"
  #   actions = ["sts:AssumeRole"]

  #   principals {
  #     identifiers = ["arn:aws:iam::${data.aws_caller_identity.global.account_id}:root"]
  #     type        = "AWS"
  #   }
  # }

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
