terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.40.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.9.0"
    }
    local = {
      source = "hashicorp/local"
    }
  }
  required_version = "1.7.4"
}
