terraform {
  required_version = ">= 1.5.2"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.82.0"
      configuration_aliases = [
        aws.region,
        aws.management,
      ]
    }
  }
}
