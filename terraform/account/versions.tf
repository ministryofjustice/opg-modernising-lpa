terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.91.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.22.0"
    }
  }
  required_version = "1.11.2"
}
