terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.86.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.20.0"
    }
  }
  required_version = "1.10.5"
}
