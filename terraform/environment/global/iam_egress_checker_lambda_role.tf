resource "aws_iam_role" "egress_checker_lambda" {
  name               = "egress-checker-${data.aws_default_tags.current.tags.environment-name}"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume.json
  lifecycle {
    create_before_destroy = true
  }
  provider = aws.global
}
