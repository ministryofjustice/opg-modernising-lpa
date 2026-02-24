terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.33.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.31.0"
    }
  }
  required_version = "1.14.5"
}
