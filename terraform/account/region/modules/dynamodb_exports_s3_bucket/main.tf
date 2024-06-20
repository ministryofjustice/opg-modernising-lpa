resource "aws_s3_bucket" "bucket" {
  bucket        = "dynamodb-exports-${data.aws_default_tags.current.tags.application}-${data.aws_default_tags.current.tags.account-name}-${data.aws_region.current.name}"
  force_destroy = true
  provider      = aws.region
}

resource "aws_s3_bucket_ownership_controls" "bucket_object_ownership" {
  bucket = aws_s3_bucket.bucket.id

  rule {
    object_ownership = "BucketOwnerEnforced"
  }
  provider = aws.region
}

resource "aws_s3_bucket_versioning" "bucket_versioning" {
  bucket = aws_s3_bucket.bucket.id

  versioning_configuration {
    status = "Enabled"
  }
  provider = aws.region
}

resource "aws_s3_bucket_server_side_encryption_configuration" "bucket_encryption_configuration" {
  bucket = aws_s3_bucket.bucket.bucket

  rule {
    apply_server_side_encryption_by_default {
      kms_master_key_id = var.s3_bucket_server_side_encryption_key_id
      sse_algorithm     = "aws:kms"
    }
  }
  provider = aws.region
}


resource "aws_s3_bucket_lifecycle_configuration" "lifecycle" {
  provider = aws.region
  bucket   = aws_s3_bucket.bucket.id

  rule {
    id     = "retain-dynamodb-exports-for-30-days"
    status = "Enabled"
    expiration {
      days = 30
    }

  }
}

resource "aws_s3_bucket_public_access_block" "public_access_policy" {
  bucket                  = aws_s3_bucket.bucket.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
  provider                = aws.region
}

resource "aws_s3_bucket_policy" "bucket" {
  depends_on = [aws_s3_bucket_public_access_block.public_access_policy]
  bucket     = aws_s3_bucket.bucket.id
  policy     = data.aws_iam_policy_document.bucket.json
  provider   = aws.region
}

resource "aws_s3_bucket_logging" "bucket" {
  bucket        = aws_s3_bucket.bucket.id
  target_bucket = data.aws_s3_bucket.access_logging.id
  target_prefix = "log/${aws_s3_bucket.bucket.id}/"
  provider      = aws.region
}

data "aws_iam_policy_document" "bucket" {
  policy_id = "PutObjPolicy"

  statement {
    sid       = "DenyUnEncryptedObjectUploads"
    effect    = "Deny"
    actions   = ["s3:PutObject"]
    resources = ["${aws_s3_bucket.bucket.arn}/*"]

    condition {
      test     = "StringNotEquals"
      variable = "s3:x-amz-server-side-encryption"
      values   = ["aws:kms"]
    }

    principals {
      type        = "AWS"
      identifiers = ["*"]
    }
  }

  statement {
    sid     = "DenyNonSSLRequests"
    effect  = "Deny"
    actions = ["s3:*"]
    resources = [
      aws_s3_bucket.bucket.arn,
      "${aws_s3_bucket.bucket.arn}/*"
    ]

    condition {
      test     = "Bool"
      variable = "aws:SecureTransport"
      values   = [false]
    }

    principals {
      type        = "AWS"
      identifiers = ["*"]
    }
  }

  provider = aws.region
}
