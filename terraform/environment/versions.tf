terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.12.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.28.2"
    }
  }
  required_version = "1.13.1"
}
