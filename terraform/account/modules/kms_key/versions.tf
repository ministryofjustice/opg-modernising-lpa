terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.100.0"
      configuration_aliases = [
        aws.eu_west_1,
        aws.eu_west_2,
      ]
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.27.1"
    }
  }
  required_version = "1.12.2"
}
