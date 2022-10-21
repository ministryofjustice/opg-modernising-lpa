data "aws_s3_bucket" "access_log_bucket" {
  bucket   = "s3-access-logs-opg-modernising-lpa-${data.aws_default_tags.current.tags.account-name}-${data.aws_region.current.name}"
  provider = aws.region
}
