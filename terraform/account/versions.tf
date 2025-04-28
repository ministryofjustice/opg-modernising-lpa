terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.96.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.24.2"
    }
  }
  required_version = "1.11.4"
}
