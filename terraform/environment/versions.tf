terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.87.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.20.0"
    }
    local = {
      source = "hashicorp/local"
    }
  }
  required_version = "1.10.5"
}
