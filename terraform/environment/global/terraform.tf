terraform {
  required_version = ">= 1.2.2"

  required_providers {
    aws = {
      source = "hashicorp/aws"
      configuration_aliases = [
        aws.global,
      ]
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "~> 2.15.0"
    }
  }
}

data "aws_default_tags" "current" {
  provider = aws.global
}

data "aws_caller_identity" "global" {
  provider = aws.global
}

data "aws_region" "global" {
  provider = aws.global
}
