resource "aws_iam_role" "cross_account_put" {
  name               = "${data.aws_default_tags.current.tags.environment-name}-cross-account-put"
  assume_role_policy = data.aws_iam_policy_document.cross_account_put_assume_role.json
  provider           = aws.global
}

data "aws_iam_policy_document" "cross_account_put_assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["events.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
  provider = aws.global
}
