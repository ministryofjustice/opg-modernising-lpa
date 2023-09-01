terraform {
  required_version = ">= 1.5.2"

  required_providers {
    aws = {
      source = "hashicorp/aws"
      configuration_aliases = [
        aws.region,
      ]
    }
  }
}
