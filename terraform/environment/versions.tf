terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.8.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.27.3"
    }
  }
  required_version = "1.12.2"
}
