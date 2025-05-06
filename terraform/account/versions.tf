terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.97.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.25.0"
    }
  }
  required_version = "1.11.4"
}
