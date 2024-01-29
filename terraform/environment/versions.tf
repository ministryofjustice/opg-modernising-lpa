terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.33.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.5.2"
    }
    local = {
      source = "hashicorp/local"
    }
  }
  required_version = "1.7.1"
}
