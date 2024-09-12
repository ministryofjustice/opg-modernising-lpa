resource "aws_s3_bucket_lifecycle_configuration" "main" {
  bucket = aws_s3_bucket.bucket.id

  rule {
    id = "delete-objects-with-viruses"

    filter {
      tag {
        key   = "virus-scan-status"
        value = "infected"
      }
    }

    expiration {
      days = "1"
    }

    status = "Enabled"
  }

  rule {
    id = "delete-objects-replicated-to-sirius"

    filter {
      tag {
        key   = "replicate"
        value = "true"
      }
    }

    expiration {
      days = "30"
    }

    status = "Enabled"
  }

  rule {
    id = "abort-incomplete-multipart-upload"

    abort_incomplete_multipart_upload {
      days_after_initiation = 7
    }

    status = "Enabled"
  }

  provider = aws.region
}
