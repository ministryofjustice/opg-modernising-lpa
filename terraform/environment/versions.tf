terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.54.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.13.1"
    }
    local = {
      source = "hashicorp/local"
    }
  }
  required_version = "1.8.5"
}
