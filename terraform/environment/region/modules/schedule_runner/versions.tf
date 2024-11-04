terraform {
  required_version = ">= 1.5.2"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.74.0"
      configuration_aliases = [
        aws.region,
        aws.management
      ]
    }
  }
}
