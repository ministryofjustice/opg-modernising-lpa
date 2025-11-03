terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.19.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.30.4"
    }
  }
  required_version = "1.13.4"
}
