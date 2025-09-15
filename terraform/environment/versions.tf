terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.13.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.29.0"
    }
  }
  required_version = "1.13.2"
}
