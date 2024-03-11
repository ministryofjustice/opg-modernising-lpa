data "aws_caller_identity" "current" {
  provider = aws.region
}

data "aws_kms_key_alias" "lambda" {
  name     = "aws/lambda"
  provider = aws.region
}
