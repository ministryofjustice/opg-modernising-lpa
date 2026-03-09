terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.35.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.31.3"
    }
  }
  required_version = "1.14.6"
}
