terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.59.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.15.0"
    }
  }
  required_version = "1.9.3"
}
