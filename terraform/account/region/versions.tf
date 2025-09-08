terraform {
  required_version = ">= 1.5.2"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.12.0"
      configuration_aliases = [
        aws.region,
        aws.management,
        aws.global,
        aws.shared,
      ]
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.28.2"
    }
  }
}
