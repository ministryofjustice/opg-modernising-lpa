terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.84.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.19.2"
    }
  }
  required_version = "1.10.4"
}
