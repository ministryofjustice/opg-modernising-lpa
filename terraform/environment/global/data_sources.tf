data "aws_default_tags" "current" {
  provider = aws.global
}

data "aws_caller_identity" "global" {
  provider = aws.global
}
