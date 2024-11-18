terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.76.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.17.2"
    }
  }
  required_version = "1.9.8"
}
