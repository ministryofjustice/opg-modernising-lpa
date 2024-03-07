data "aws_default_tags" "current" {
  provider = aws.global
}

data "aws_caller_identity" "global" {
  provider = aws.global
}

data "aws_iam_policy_document" "lambda_assume" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
  provider = aws.global
}
