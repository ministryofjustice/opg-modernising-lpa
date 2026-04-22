terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.40.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.32.1"
    }
  }
  required_version = "1.14.8"
}
