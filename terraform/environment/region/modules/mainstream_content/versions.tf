terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.23.0"
      configuration_aliases = [
        aws.region,
      ]
    }
    local = {
      source = "hashicorp/local"
    }
  }

  required_version = ">= 1.9.8"
}
