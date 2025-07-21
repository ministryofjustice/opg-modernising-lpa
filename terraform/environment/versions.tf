terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.100.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.27.1"
    }
  }
  required_version = "1.12.2"
}
