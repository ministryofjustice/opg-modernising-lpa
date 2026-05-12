terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.43.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.32.4"
    }
  }
  required_version = "1.15.1"
}
