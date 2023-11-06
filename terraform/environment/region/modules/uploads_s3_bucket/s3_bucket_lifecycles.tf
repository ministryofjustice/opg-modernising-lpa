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

  provider = aws.region
}
