data "aws_caller_identity" "eu_west_1" {
  provider = aws.eu_west_1
}

data "aws_region" "eu_west_1" {
  provider = aws.eu_west_1
}
