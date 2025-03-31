terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.93.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.23.1"
    }
  }
  required_version = "1.11.3"
}
