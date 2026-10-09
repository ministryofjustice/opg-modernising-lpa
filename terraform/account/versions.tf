terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.67.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.36.0"
    }
  }
  required_version = "1.16.4"
}
