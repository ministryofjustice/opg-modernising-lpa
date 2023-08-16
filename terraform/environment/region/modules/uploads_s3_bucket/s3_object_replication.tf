data "aws_iam_role" "replication" {
  name     = "reduced-fees-uploads-replication"
  provider = aws.region
}

data "aws_iam_policy_document" "replication" {

  statement {
    effect = "Allow"

    actions = [
      "kms:Decrypt",
      "kms:GenerateDataKey",
    ]

    resources = [
      data.aws_kms_alias.reduced_fees_uploads_s3_encryption.target_key_arn,
    ]
  }
  statement {
    effect = "Allow"
    actions = [
      "kms:GenerateDataKey",
      "kms:Encrypt"
    ]
    resources = [
      var.s3_replication.destination_encryption_key_arn
    ]
  }

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
      # "s3:PutObject",
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
      "arn:aws:s3:::batch-manifests-${data.aws_default_tags.current.tags.application}-${data.aws_default_tags.current.tags.account-name}-${data.aws_region.current.name}/*"
    ]
  }
  statement {
    effect = "Allow"

    actions = [
      "s3:PutObject"
    ]

    resources = [
      "arn:aws:s3:::batch-manifests-${data.aws_default_tags.current.tags.application}-${data.aws_default_tags.current.tags.account-name}-${data.aws_region.current.name}/*"
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


resource "aws_ssm_parameter" "s3_batch_configuration" {
  name = "/modernising-lpa/s3-batch-configuration/${data.aws_default_tags.current.tags.environment-name}/s3_batch_configuration"
  type = "String"
  value = jsonencode({
    "aws_account_id" : data.aws_caller_identity.current.account_id,
    "report_and_manifests_bucket" : "arn:aws:s3:::batch-manifests-${data.aws_default_tags.current.tags.application}-${data.aws_default_tags.current.tags.account-name}-${data.aws_region.current.name}",
    "source_bucket" : aws_s3_bucket.bucket.arn,
    "role_arn" : data.aws_iam_role.replication.arn,
    "aws_region" : data.aws_region.current.name,
  })
  provider = aws.region
}
