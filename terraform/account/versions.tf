terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.52.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.12.2"
    }
  }
  required_version = "1.8.5"
}
