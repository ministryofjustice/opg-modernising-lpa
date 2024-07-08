terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.57.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.14.5"
    }
    local = {
      source = "hashicorp/local"
    }
  }
  required_version = "1.9.1"
}
