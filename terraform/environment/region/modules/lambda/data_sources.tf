data "aws_caller_identity" "current" {
  provider = aws.region
}

data "aws_kms_alias" "lambda" {
  name     = "alias/aws/lambda"
  provider = aws.region
}
