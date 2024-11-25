terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.77.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.18.1"
    }
  }
  required_version = "1.9.8"
}
