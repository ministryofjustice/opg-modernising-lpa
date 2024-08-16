terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.62.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.15.2"
    }
  }
  required_version = "1.9.4"
}
