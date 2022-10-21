resource "aws_s3_bucket" "access_log" {
  provider = aws.region
  bucket   = "${data.aws_default_tags.current.tags.application}-${data.aws_default_tags.current.tags.account-name}-lb-access-logs-${data.aws_region.current.name}"
}

resource "aws_s3_bucket_acl" "access_log" {
  provider = aws.region
  bucket   = aws_s3_bucket.access_log.id
  acl      = "private"
}

resource "aws_s3_bucket_server_side_encryption_configuration" "access_log" {
  provider = aws.region
  bucket   = aws_s3_bucket.access_log.id
  rule {
    apply_server_side_encryption_by_default {
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

data "aws_iam_policy_document" "access_log" {
  provider = aws.region
  statement {
    sid = "accessLogBucketAccess"
    resources = [
      aws_s3_bucket.access_log.arn,
      "${aws_s3_bucket.access_log.arn}/*",
    ]
    effect  = "Allow"
    actions = ["s3:PutObject"]
    principals {
      identifiers = [data.aws_elb_service_account.main.id]
      type        = "AWS"
    }
  }

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
}

resource "aws_s3_bucket_public_access_block" "access_log" {
  provider                = aws.region
  bucket                  = aws_s3_bucket.access_log.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}
