terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 3.59.0"
    }
  }
}

data "aws_region" "current" {}

data "aws_caller_identity" "current" {}
