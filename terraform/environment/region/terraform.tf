terraform {
  required_version = ">= 1.5.2"

  required_providers {
    aws = {
      source = "hashicorp/aws"
      configuration_aliases = [
        aws.region,
        aws.global,
        aws.management_global,
      ]
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "~> 2.15.0"
    }
  }
}

data "aws_region" "current" {
  provider = aws.region
}

data "aws_caller_identity" "current" {
  provider = aws.region
}

data "aws_default_tags" "current" {
  provider = aws.region
}

data "aws_caller_identity" "global" {
  provider = aws.global
}

data "aws_region" "global" {
  provider = aws.global
}
