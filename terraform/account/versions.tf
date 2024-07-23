terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.59.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.14.5"
    }
  }
  required_version = "1.9.2"
}
