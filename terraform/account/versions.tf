terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.82.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.18.1"
    }
  }
  required_version = "1.10.3"
}
