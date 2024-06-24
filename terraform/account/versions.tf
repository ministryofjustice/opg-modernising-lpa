terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.55.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.14.3"
    }
  }
  required_version = "1.8.5"
}
