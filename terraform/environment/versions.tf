terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.16.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.30.2"
    }
  }
  required_version = "1.13.3"
}
