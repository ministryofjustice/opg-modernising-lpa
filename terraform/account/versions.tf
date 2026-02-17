terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.32.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.30.9"
    }
  }
  required_version = "1.14.5"
}
