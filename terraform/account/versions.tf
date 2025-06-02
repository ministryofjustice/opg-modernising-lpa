terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.99.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.26.0"
    }
  }
  required_version = "1.12.1"
}
