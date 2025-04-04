terraform {
  required_version = ">= 1.5.2"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.93.0"
      configuration_aliases = [
        aws.region,
        aws.management,
        aws.global,
      ]
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.23.1"
    }
  }
}
