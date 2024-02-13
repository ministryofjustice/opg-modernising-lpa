terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.35.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.7.0"
    }
    local = {
      source = "hashicorp/local"
    }
  }
  required_version = "1.7.2"
}
