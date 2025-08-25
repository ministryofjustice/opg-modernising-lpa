terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.10.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.28.1"
    }
  }
  required_version = "1.13.0"
}
