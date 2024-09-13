resource "aws_s3_bucket" "access_log" {
  provider = aws.region
  bucket   = "${data.aws_default_tags.current.tags.application}-${data.aws_default_tags.current.tags.account-name}-lb-access-logs-${data.aws_region.current.name}"
}

#tfsec:ignore:aws-s3-encryption-customer-key
resource "aws_s3_bucket_server_side_encryption_configuration" "access_log" {
  provider = aws.region
  bucket   = aws_s3_bucket.access_log.id
  rule {
    apply_server_side_encryption_by_default {
      # Access logging for ALBs only supports Amazon S3 Managed Encryption Keys (SSE-S3)
      sse_algorithm = "aws:kms"
    }
  }
}

resource "aws_s3_bucket_versioning" "access_log" {
  provider = aws.region
  bucket   = aws_s3_bucket.access_log.id
  versioning_configuration {
    status = "Enabled"
  }
}

data "aws_s3_bucket" "s3_access_logging" {
  provider = aws.region
  bucket   = "s3-access-logs-${data.aws_default_tags.current.tags.application}-${data.aws_default_tags.current.tags.account-name}-${data.aws_region.current.name}"
}

resource "aws_s3_bucket_logging" "access_log" {
  provider = aws.region
  bucket   = aws_s3_bucket.access_log.id

  target_bucket = data.aws_s3_bucket.s3_access_logging.id
  target_prefix = "lb-access-log/"
}

resource "aws_s3_bucket_policy" "access_log" {
  provider = aws.region
  bucket   = aws_s3_bucket.access_log.id
  policy   = data.aws_iam_policy_document.access_log.json
}

data "aws_elb_service_account" "main" {
  provider = aws.region
  region   = data.aws_region.current.name
}

resource "aws_s3_bucket_lifecycle_configuration" "lifecycle" {
  provider = aws.region
  bucket   = aws_s3_bucket.access_log.id

  rule {
    id     = "retain-dynamodb-exports-for-400-days"
    status = "Enabled"
    expiration {
      days = 400
    }
    noncurrent_version_expiration {
      noncurrent_days = 400
    }
  }
  rule {
    id     = "abort-incomplete-multipart-upload"
    status = "Enabled"
    abort_incomplete_multipart_upload {
      days_after_initiation = 7
    }

  }
}

data "aws_iam_policy_document" "access_log" {
  provider = aws.region

  statement {
    sid = "accessLogDelivery"
    resources = [
      aws_s3_bucket.access_log.arn,
      "${aws_s3_bucket.access_log.arn}/*",
    ]
    effect  = "Allow"
    actions = ["s3:PutObject"]
    principals {
      identifiers = ["delivery.logs.amazonaws.com"]
      type        = "Service"
    }
    condition {
      test     = "StringEquals"
      values   = ["bucket-owner-full-control"]
      variable = "s3:x-amz-acl"
    }
  }

  statement {
    sid = "accessLogBucketAccess"
    resources = [
      aws_s3_bucket.access_log.arn,
      "${aws_s3_bucket.access_log.arn}/*",
      "${aws_s3_bucket.access_log.arn}/*/AWSLogs/${data.aws_caller_identity.current.account_id}/*",
    ]
    effect  = "Allow"
    actions = ["s3:PutObject"]
    principals {
      identifiers = [data.aws_elb_service_account.main.id]
      type        = "AWS"
    }
  }

  statement {
    sid = "accessGetAcl"
    resources = [
      aws_s3_bucket.access_log.arn
    ]
    effect  = "Allow"
    actions = ["s3:GetBucketAcl"]
    principals {
      identifiers = ["delivery.logs.amazonaws.com"]
      type        = "Service"
    }
  }

  statement {
    sid     = "AllowSSLRequestsOnly"
    effect  = "Deny"
    actions = ["s3:*"]
    resources = [
      aws_s3_bucket.access_log.arn,
      "${aws_s3_bucket.access_log.arn}/*",
    ]
    condition {
      test     = "Bool"
      values   = ["false"]
      variable = "aws:SecureTransport"
    }
    principals {
      identifiers = ["*"]
      type        = "AWS"
    }
  }

  # statement {
  #   sid = "DenyUnEncryptedObjectUploads"
  #   resources = [
  #     aws_s3_bucket.access_log.arn,
  #     "${aws_s3_bucket.access_log.arn}/*",
  #   ]
  #   effect  = "Deny"
  #   actions = ["s3:PutObject"]
  #   principals {
  #     identifiers = ["*"]
  #     type        = "AWS"
  #   }
  #   condition {
  #     test     = "StringNotEquals"
  #     values   = ["aws:kms"]
  #     variable = "s3:x-amz-server-side-encryption"
  #   }
  # }
}

resource "aws_s3_bucket_public_access_block" "access_log" {
  provider                = aws.region
  bucket                  = aws_s3_bucket.access_log.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_lifecycle_configuration" "log_retention_policy" {
  provider = aws.region
  bucket   = aws_s3_bucket.access_log.id

  rule {
    id     = "retain-logs-for-13-months"
    status = "Enabled"

    transition {
      days          = 30
      storage_class = "STANDARD_IA"
    }

    transition {
      days          = 60
      storage_class = "GLACIER"
    }

    expiration {
      days = 400
    }

  }
}

data "aws_iam_role" "sns_success_feedback" {
  name     = "SNSSuccessFeedback"
  provider = aws.global
}

data "aws_iam_role" "sns_failure_feedback" {
  provider = aws.global
  name     = "SNSFailureFeedback"
}


module "s3_event_notifications" {
  providers = { aws = aws.region }
  source    = "./modules/s3_bucket_event_notifications"
  s3_bucket_event_types = [
    "s3:ObjectRemoved:*",
    "s3:ObjectAcl:Put",
  ]
  sns_kms_key_alias             = var.sns_kms_key_alias
  s3_bucket_id                  = aws_s3_bucket.access_log.id
  sns_failure_feedback_role_arn = data.aws_iam_role.sns_failure_feedback.arn
  sns_success_feedback_role_arn = data.aws_iam_role.sns_success_feedback.arn
}
