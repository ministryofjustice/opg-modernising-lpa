terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.56.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.14.4"
    }
  }
  required_version = "1.9.0"
}
