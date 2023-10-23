resource "aws_s3_bucket_notification" "bucket_notification" {
  bucket = aws_s3_bucket.bucket.id

  lambda_function {
    id                  = "bucket-av-scan"
    lambda_function_arn = var.s3_antivirus_lambda_function.arn
    events              = ["s3:ObjectCreated:Put"]
  }

  lambda_function {
    id                  = "av-object-tagging"
    lambda_function_arn = var.events_received_lambda_function.arn
    events              = ["s3:ObjectTagging:Put"]
  }
  depends_on = [aws_lambda_permission.av_scan, aws_lambda_permission.object_tagging]
  provider   = aws.region
}

resource "aws_lambda_permission" "av_scan" {
  statement_id   = "AllowExecutionFromS3BucketAvScan"
  action         = "lambda:InvokeFunction"
  function_name  = var.s3_antivirus_lambda_function.function_name
  principal      = "s3.amazonaws.com"
  source_account = data.aws_caller_identity.current.account_id
  source_arn     = aws_s3_bucket.bucket.arn
  provider       = aws.region
}

resource "aws_lambda_permission" "object_tagging" {
  statement_id   = "AllowExecutionFromS3BucketObjectTagging"
  action         = "lambda:InvokeFunction"
  function_name  = var.events_received_lambda_function.function_name
  principal      = "s3.amazonaws.com"
  source_account = data.aws_caller_identity.current.account_id
  source_arn     = aws_s3_bucket.bucket.arn
  provider       = aws.region
}
