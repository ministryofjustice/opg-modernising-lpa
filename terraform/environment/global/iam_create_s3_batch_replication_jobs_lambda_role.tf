resource "aws_iam_role" "create_s3_batch_replication_jobs_lambda_role" {
  name               = "create-s3-batch-replication-jobs-${data.aws_default_tags.current.tags.environment-name}"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume.json
  lifecycle {
    create_before_destroy = true
  }
  provider = aws.global
}
