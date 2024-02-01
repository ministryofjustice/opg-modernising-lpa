data "aws_iam_policy_document" "assume_replication_role" {
  statement {
    effect = "Allow"

    principals {
      type = "Service"
      identifiers = [
        "s3.amazonaws.com",
        "batchoperations.s3.amazonaws.com"
      ]
    }

    actions = ["sts:AssumeRole"]
  }
  provider = aws.global
}

resource "aws_iam_role" "replication" {
  name               = "reduced-fees-uploads-replication"
  assume_role_policy = data.aws_iam_policy_document.assume_replication_role.json
  provider           = aws.global
}
