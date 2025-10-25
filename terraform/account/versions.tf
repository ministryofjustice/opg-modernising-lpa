terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.17.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.30.3"
    }
  }
  required_version = "1.13.4"
}
