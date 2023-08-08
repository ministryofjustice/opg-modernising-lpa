data "aws_iam_role" "replication" {
  name     = "reduced-fees-uploads-replication"
  provider = aws.region
}

data "aws_iam_policy_document" "replication" {
  statement {
    effect = "Allow"

    actions = [
      "s3:GetReplicationConfiguration",
      "s3:ListBucket",
      "s3:PutInventoryConfiguration",
    ]

    resources = [aws_s3_bucket.bucket.arn]
  }

  statement {
    effect = "Allow"

    actions = [
      "s3:GetObjectVersionForReplication",
      "s3:GetObjectVersionAcl",
      "s3:GetObjectVersionTagging",
      "s3:InitiateReplication",
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

    resources = ["${var.s3_replication.destination_bucket_arn}/*"]
  }
  statement {
    actions = [
      "s3:GetObject",
      "s3:GetObjectVersion"
    ]

    effect = "Allow"

    resources = [
      "arn:aws:s3:::replication-manifest-opg-modernising-lpa-605mlpab119-eu-west-1/*"
    ]
  }
  statement {
    effect = "Allow"

    actions = [
      "s3:PutObject"
    ]

    resources = [
      "arn:aws:s3:::replication-manifest-opg-modernising-lpa-605mlpab119-eu-west-1/*"
    ]
  }
  provider = aws.region
}

resource "aws_iam_policy" "replication" {
  name     = "reduced-fees-uploads-replication-${data.aws_default_tags.current.tags.environment-name}-${data.aws_region.current.name}"
  policy   = data.aws_iam_policy_document.replication.json
  provider = aws.region
}

resource "aws_iam_role_policy_attachment" "replication" {
  role       = data.aws_iam_role.replication.name
  policy_arn = aws_iam_policy.replication.arn
  provider   = aws.region
}

resource "aws_s3_bucket_replication_configuration" "replication" {
  depends_on = [aws_s3_bucket_versioning.bucket_versioning]
  role       = data.aws_iam_role.replication.arn
  bucket     = aws_s3_bucket.bucket.id

  rule {
    id = "whenScannedOkAndReadyToReplicate"

    source_selection_criteria {
      replica_modifications {
        status = "Disabled"
      }
      sse_kms_encrypted_objects {
        status = "Enabled"
      }
    }

    delete_marker_replication {
      status = "Disabled"
    }
    filter {
      and {
        tags = {
          "replicate"         = "true"
          "virus-scan-status" = "ok"
        }
      }
    }

    status = var.s3_replication.enabled ? "Enabled" : "Disabled"

    destination {
      account = var.s3_replication.destination_account_id
      bucket  = var.s3_replication.destination_bucket_arn

      access_control_translation {
        owner = "Destination"
      }

      encryption_configuration {
        replica_kms_key_id = var.s3_replication.destination_encryption_key_arn
      }

      metrics {
        event_threshold {
          minutes = 15
        }
        status = "Enabled"
      }

      replication_time {
        status = "Enabled"
        time {
          minutes = 15
        }
      }
    }
  }
  provider = aws.region
}
