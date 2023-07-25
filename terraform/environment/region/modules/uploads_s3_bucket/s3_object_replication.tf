
data "aws_iam_policy_document" "assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["s3.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
  provider = aws.region
}

resource "aws_iam_role" "replication" {
  name               = "reduced-fees-uploads-replication-${data.aws_default_tags.current.tags.environment-name}-${data.aws_region.current.name}"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
  provider           = aws.region
}

data "aws_iam_policy_document" "replication" {
  statement {
    effect = "Allow"

    actions = [
      "s3:GetReplicationConfiguration",
      "s3:ListBucket",
    ]

    resources = [aws_s3_bucket.bucket.arn]
  }

  statement {
    effect = "Allow"

    actions = [
      "s3:GetObjectVersionForReplication",
      "s3:GetObjectVersionAcl",
      "s3:GetObjectVersionTagging",
    ]

    resources = ["${aws_s3_bucket.bucket.arn}/*"]
  }

  statement {
    effect = "Allow"

    actions = [
      "s3:ReplicateObject",
      "s3:ReplicateDelete",
      "s3:ReplicateTags",
    ]

    resources = ["${var.s3_replication_target_bucket_arn}/*"]
  }
  provider = aws.region
}

resource "aws_iam_policy" "replication" {
  name     = "reduced_fees_uploads_replication"
  policy   = data.aws_iam_policy_document.replication.json
  provider = aws.region
}

resource "aws_iam_role_policy_attachment" "replication" {
  role       = aws_iam_role.replication.name
  policy_arn = aws_iam_policy.replication.arn
  provider   = aws.region
}

resource "aws_s3_bucket_replication_configuration" "replication" {
  depends_on = [aws_s3_bucket_versioning.bucket_versioning]
  role       = aws_iam_role.replication.arn
  bucket     = aws_s3_bucket.bucket.id

  rule {
    id = "whenScannedOkAndReadyToReplicate"

    delete_marker_replication {
      status = "Disabled"
    }

    filter {
      tag {
        key   = "replicate"
        value = "true"
      }
      and {
        tags = {
          key   = "virus-scan-status"
          value = "ok"
        }
      }
    }

    status = "Enabled"

    destination {
      bucket        = var.s3_replication_target_bucket_arn
      storage_class = "STANDARD"
    }
  }
  provider = aws.region
}
