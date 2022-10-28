resource "aws_s3_bucket_notification" "bucket_notification" {
  bucket = var.s3_bucket_id
  topic {
    topic_arn     = aws_sns_topic.s3_event_notification.arn
    events        = var.s3_bucket_event_types
    filter_suffix = ".log"
  }
}
