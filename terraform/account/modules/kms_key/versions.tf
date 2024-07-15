terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.58.0"
      configuration_aliases = [
        aws.eu_west_1,
        aws.eu_west_2,
      ]
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.14.5"
    }
  }
  required_version = "1.9.2"
}
