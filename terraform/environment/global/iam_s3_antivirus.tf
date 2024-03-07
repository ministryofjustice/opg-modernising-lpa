resource "aws_iam_role" "s3_antivirus" {
  name               = "s3-antivirus-${data.aws_default_tags.current.tags.environment-name}"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume.json
  provider           = aws.global
}

resource "aws_iam_role_policy_attachment" "s3_antivirus_execution_role" {
  role       = aws_iam_role.s3_antivirus.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"
  provider   = aws.global
}
