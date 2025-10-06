terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.15.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.30.0"
    }
  }
  required_version = "1.13.3"
}
