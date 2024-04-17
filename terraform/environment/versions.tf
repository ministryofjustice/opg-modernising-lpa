terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.45.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.11.3"
    }
    local = {
      source = "hashicorp/local"
    }
  }
  required_version = "1.8.0"
}
