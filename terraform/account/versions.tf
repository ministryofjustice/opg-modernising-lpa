terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.64.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.15.6"
    }
  }
  required_version = "1.9.5"
}
