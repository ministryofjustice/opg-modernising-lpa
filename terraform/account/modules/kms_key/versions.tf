terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.67.0"
      configuration_aliases = [
        aws.eu_west_1,
        aws.eu_west_2,
      ]
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.36.0"
    }
  }
  required_version = "1.16.5"
}
