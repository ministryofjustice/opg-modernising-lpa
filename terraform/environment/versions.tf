terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.52.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.33.0"
    }
  }
  required_version = "1.15.7"
}
