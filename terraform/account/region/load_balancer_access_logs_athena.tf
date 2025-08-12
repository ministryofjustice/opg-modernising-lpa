resource "aws_s3_bucket" "athena_results" {
  count         = var.athena_enabled ? 1 : 0
  bucket        = "${data.aws_default_tags.current.tags.application}-${data.aws_default_tags.current.tags.account-name}-lb-logs-athena-${data.aws_region.current.region}"
  force_destroy = true
  provider      = aws.region
}

resource "aws_s3_bucket_ownership_controls" "athena_results" {
  count  = var.athena_enabled ? 1 : 0
  bucket = aws_s3_bucket.athena_results[0].id
  rule {
    object_ownership = "BucketOwnerEnforced"
  }
  provider = aws.region
}

resource "aws_s3_bucket_versioning" "versioning_example" {
  count  = var.athena_enabled ? 1 : 0
  bucket = aws_s3_bucket.athena_results[0].id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_lifecycle_configuration" "athena_results" {
  count  = var.athena_enabled ? 1 : 0
  bucket = aws_s3_bucket.athena_results[0].id

  rule {
    id     = "ExpireObjectsAfter28Days"
    status = "Enabled"

    filter {
      prefix = ""
    }

    expiration {
      days = 28
    }
  }
  provider = aws.region
}

resource "aws_s3_bucket_server_side_encryption_configuration" "athena_results" {
  count  = var.athena_enabled ? 1 : 0
  bucket = aws_s3_bucket.athena_results[0].id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm     = "aws:kms"
      kms_master_key_id = var.athena_s3_target_key_id
    }
  }
  provider = aws.region
}

resource "aws_s3_bucket_public_access_block" "athena_results" {
  count  = var.athena_enabled ? 1 : 0
  bucket = aws_s3_bucket.athena_results[0].id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
  provider                = aws.region
}

resource "aws_s3_bucket_policy" "athena_results" {
  count      = var.athena_enabled ? 1 : 0
  depends_on = [aws_s3_bucket_public_access_block.athena_results[0]]
  bucket     = aws_s3_bucket.athena_results[0].id
  policy     = data.aws_iam_policy_document.athena_results[0].json
  provider   = aws.region
}

resource "aws_s3_bucket_logging" "athena_results" {
  count  = var.athena_enabled ? 1 : 0
  bucket = aws_s3_bucket.athena_results[0].id

  target_bucket = aws_s3_bucket.access_log.id
  target_prefix = "log/${aws_s3_bucket.athena_results[0].id}/"
  provider      = aws.region
}

data "aws_iam_policy_document" "athena_results" {
  count     = var.athena_enabled ? 1 : 0
  policy_id = "PutObjPolicy"

  statement {
    sid     = "DenyNoneSSLRequests"
    effect  = "Deny"
    actions = ["s3:*"]
    resources = [
      aws_s3_bucket.athena_results[0].arn,
      "${aws_s3_bucket.athena_results[0].arn}/*"
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

  statement {
    sid     = "AllowOperatorAccess"
    effect  = "Allow"
    actions = ["s3:*"]
    resources = [
      aws_s3_bucket.athena_results[0].arn,
      "${aws_s3_bucket.athena_results[0].arn}/*"
    ]

    principals {
      type        = "AWS"
      identifiers = ["arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/operator"]
    }
  }
  provider = aws.region
}

resource "aws_athena_workgroup" "alb_logs" {
  count         = var.athena_enabled ? 1 : 0
  name          = "${data.aws_default_tags.current.tags.account-name}-${data.aws_region.current.region}"
  description   = "Workgroup for the interrogation of Load Balancer Logs in ${data.aws_default_tags.current.tags.account-name} ${data.aws_region.current.region}"
  force_destroy = true

  configuration {
    enforce_workgroup_configuration    = true
    publish_cloudwatch_metrics_enabled = true

    result_configuration {
      output_location = "s3://${aws_s3_bucket.athena_results[0].bucket}/workspace/"

      encryption_configuration {
        encryption_option = "SSE_S3"
      }
    }
  }
  provider = aws.region
}

resource "aws_athena_database" "access_logs" {
  count         = var.athena_enabled ? 1 : 0
  name          = "load_balancer_logs"
  bucket        = aws_s3_bucket.athena_results[0].id
  force_destroy = true

  encryption_configuration {
    encryption_option = "SSE_S3"
  }
  provider = aws.region
}

resource "aws_athena_named_query" "create_alb_log_table" {
  count       = var.athena_enabled ? 1 : 0
  name        = "create-alb-log-table"
  description = "Query to create the ALB Logging Table for an Environment"
  workgroup   = aws_athena_workgroup.alb_logs[0].id
  database    = aws_athena_database.access_logs[0].name
  query       = templatefile("${path.module}/load_balancer_logs_create_table.tpl", local.template_vars)
  provider    = aws.region
}

locals {
  template_vars = {
    bucket     = aws_s3_bucket.access_log.id
    account_id = data.aws_caller_identity.current.account_id
    region     = data.aws_region.current.region
    workspace  = data.aws_default_tags.current.tags.account-name
  }
}
