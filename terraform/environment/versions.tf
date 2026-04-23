terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.41.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.32.2"
    }
  }
  required_version = "1.14.8"
}
