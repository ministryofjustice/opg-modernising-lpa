terraform {
  required_version = ">= 1.5.2"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.100.0"
      configuration_aliases = [
        aws.region,
        aws.global,
        aws.management_global,
        aws.management,
      ]
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.27.1"
    }
  }
}
