terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.61.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.15.1"
    }
    local = {
      source = "hashicorp/local"
    }
  }
  required_version = "1.9.3"
}
