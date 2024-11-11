terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.75.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.17.0"
    }
  }
  required_version = "1.9.8"
}
