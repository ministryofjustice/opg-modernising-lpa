terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.21.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.30.5"
    }
  }
  required_version = "1.14.0"
}
