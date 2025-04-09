resource "aws_iam_role" "opg_metrics" {
  name               = "s3-antivirus-${data.aws_default_tags.current.tags.environment-name}"
  assume_role_policy = data.aws_iam_policy_document.opg_metrics_assume_role.json
  provider           = aws.global
}

data "aws_iam_policy_document" "opg_metrics_assume_role" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      identifiers = ["events.amazonaws.com"]
      type        = "Service"
    }
  }
  provider = aws.global
}
