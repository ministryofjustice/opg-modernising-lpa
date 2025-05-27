terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.98.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.25.2"
    }
  }
  required_version = "1.12.1"
}
